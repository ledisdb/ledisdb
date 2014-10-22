package server

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/log"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/rpl"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	errConnectMaster = errors.New("connect master error")
)

type master struct {
	sync.Mutex

	conn net.Conn
	rb   *bufio.Reader

	app *App

	quit chan struct{}

	addr string

	wg sync.WaitGroup

	syncBuf bytes.Buffer
}

func newMaster(app *App) *master {
	m := new(master)
	m.app = app

	m.quit = make(chan struct{}, 1)

	return m
}

var (
	quitCmd = []byte("*1\r\n$4\r\nquit\r\n")
)

func (m *master) Close() {
	ledis.AsyncNotify(m.quit)

	if m.conn != nil {
		//for replication, we send quit command to close gracefully
		m.conn.Write(quitCmd)

		m.conn.Close()
		m.conn = nil
	}

	m.wg.Wait()
}

func (m *master) connect() error {
	if len(m.addr) == 0 {
		return fmt.Errorf("no assign master addr")
	}

	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}

	if conn, err := net.Dial("tcp", m.addr); err != nil {
		return err
	} else {
		m.conn = conn

		m.rb = bufio.NewReaderSize(m.conn, 4096)
	}

	return nil
}

func (m *master) stopReplication() error {
	m.Close()

	return nil
}

func (m *master) startReplication(masterAddr string, restart bool) error {
	//stop last replcation, if avaliable
	m.Close()

	m.addr = masterAddr

	m.quit = make(chan struct{}, 1)

	m.app.cfg.Readonly = true

	m.wg.Add(1)
	go m.runReplication(restart)
	return nil
}

func (m *master) runReplication(restart bool) {
	defer m.wg.Done()

	for {
		select {
		case <-m.quit:
			return
		default:
			if err := m.connect(); err != nil {
				log.Error("connect master %s error %s, try 2s later", m.addr, err.Error())
				time.Sleep(2 * time.Second)
				continue
			}
		}

		if err := m.replConf(); err != nil {
			log.Error("replconf error %s", err.Error())
			return
		}

		if restart {
			if err := m.fullSync(); err != nil {
				if m.conn != nil {
					//if conn == nil, other close the replication, not error
					log.Error("restart fullsync error %s", err.Error())
				}
				return
			}
		}

		for {
			if err := m.sync(); err != nil {
				if m.conn != nil {
					//if conn == nil, other close the replication, not error
					log.Error("sync error %s", err.Error())
				}
				return
			}

			select {
			case <-m.quit:
				return
			default:
				break
			}
		}
	}

	return
}

var (
	fullSyncCmd       = []byte("*1\r\n$8\r\nfullsync\r\n")                               //fullsync
	syncCmdFormat     = "*2\r\n$4\r\nsync\r\n$%d\r\n%s\r\n"                              //sync logid
	replconfCmdFormat = "*3\r\n$8\r\nreplconf\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n" //replconf listening-port port
)

func (m *master) replConf() error {
	_, port, err := net.SplitHostPort(m.app.cfg.Addr)
	if err != nil {
		return err
	}

	cmd := hack.Slice(fmt.Sprintf(replconfCmdFormat, len(port), port))

	if _, err := m.conn.Write(cmd); err != nil {
		return err
	}

	if s, err := ReadStatus(m.rb); err != nil {
		return err
	} else if strings.ToLower(s) != "ok" {
		return fmt.Errorf("not ok but %s", s)
	}

	return nil
}

func (m *master) fullSync() error {
	log.Info("begin full sync")

	if _, err := m.conn.Write(fullSyncCmd); err != nil {
		return err
	}

	dumpPath := path.Join(m.app.cfg.DataDir, "master.dump")
	f, err := os.OpenFile(dumpPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer os.Remove(dumpPath)

	err = ReadBulkTo(m.rb, f)
	f.Close()
	if err != nil {
		log.Error("read dump data error %s", err.Error())
		return err
	}

	if _, err = m.app.ldb.LoadDumpFile(dumpPath); err != nil {
		log.Error("load dump file error %s", err.Error())
		return err
	}

	return nil
}

func (m *master) nextSyncLogID() (uint64, error) {
	s, err := m.app.ldb.ReplicationStat()
	if err != nil {
		return 0, err
	}

	if s.LastID > s.CommitID {
		return s.LastID + 1, nil
	} else {
		return s.CommitID + 1, nil
	}
}

func (m *master) sync() error {
	var err error
	var syncID uint64
	if syncID, err = m.nextSyncLogID(); err != nil {
		return err
	}

	logIDStr := strconv.FormatUint(syncID, 10)

	cmd := hack.Slice(fmt.Sprintf(syncCmdFormat, len(logIDStr),
		logIDStr))

	if _, err := m.conn.Write(cmd); err != nil {
		return err
	}

	m.syncBuf.Reset()

	if err = ReadBulkTo(m.rb, &m.syncBuf); err != nil {
		switch err.Error() {
		case ledis.ErrLogMissed.Error():
			return m.fullSync()
		case ledis.ErrRplNotSupport.Error():
			m.stopReplication()
			return nil
		default:
			return err
		}
	}

	buf := m.syncBuf.Bytes()

	if len(buf) < 8 {
		return fmt.Errorf("inavlid sync size %d", len(buf))
	}

	m.app.info.Replication.MasterLastLogID.Set(num.BytesToUint64(buf))

	var t bytes.Buffer
	m.app.info.dumpReplication(&t)

	buf = buf[8:]

	if len(buf) == 0 {
		return nil
	}

	if err = m.app.ldb.StoreLogsFromData(buf); err != nil {
		return err
	}

	return nil

}

func (app *App) slaveof(masterAddr string, restart bool, readonly bool) error {
	app.m.Lock()
	defer app.m.Unlock()

	//in master mode and no slaveof, only set readonly
	if len(app.cfg.SlaveOf) == 0 && len(masterAddr) == 0 {
		app.cfg.Readonly = readonly
		return nil
	}

	if !app.ldb.ReplicationUsed() {
		return fmt.Errorf("slaveof must enable replication")
	}

	app.cfg.SlaveOf = masterAddr

	if len(masterAddr) == 0 {
		if err := app.m.stopReplication(); err != nil {
			return err
		}

		app.cfg.Readonly = readonly
	} else {
		return app.m.startReplication(masterAddr, restart)
	}

	return nil
}

func (app *App) tryReSlaveof() error {
	app.m.Lock()
	defer app.m.Unlock()

	if !app.ldb.ReplicationUsed() {
		return nil
	}

	if len(app.cfg.SlaveOf) == 0 {
		return nil
	} else {
		return app.m.startReplication(app.cfg.SlaveOf, true)
	}
}

func (app *App) addSlave(c *client) {
	addr := c.slaveListeningAddr

	app.slock.Lock()
	defer app.slock.Unlock()

	app.slaves[addr] = c
}

func (app *App) removeSlave(c *client, activeQuit bool) {
	addr := c.slaveListeningAddr

	app.slock.Lock()
	defer app.slock.Unlock()

	if _, ok := app.slaves[addr]; ok {
		delete(app.slaves, addr)
		log.Info("remove slave %s", addr)
		if activeQuit {
			asyncNotifyUint64(app.slaveSyncAck, c.lastLogID)
		}
	}
}

func (app *App) slaveAck(c *client) {
	addr := c.slaveListeningAddr

	app.slock.Lock()
	defer app.slock.Unlock()

	if _, ok := app.slaves[addr]; !ok {
		//slave not add
		return
	}

	asyncNotifyUint64(app.slaveSyncAck, c.lastLogID)
}

func asyncNotifyUint64(ch chan uint64, v uint64) {
	select {
	case ch <- v:
	default:
	}
}

func (app *App) publishNewLog(l *rpl.Log) {
	if !app.cfg.Replication.Sync {
		//no sync replication, we will do async
		return
	}

	app.info.Replication.PubLogNum.Add(1)

	app.slock.Lock()

	slaveNum := len(app.slaves)

	total := (slaveNum + 1) / 2
	if app.cfg.Replication.WaitMaxSlaveAcks > 0 {
		total = num.MinInt(total, app.cfg.Replication.WaitMaxSlaveAcks)
	}

	n := 0
	logId := l.ID
	for _, s := range app.slaves {
		if s.lastLogID == logId {
			//slave has already owned this log
			n++
		} else if s.lastLogID > logId {
			log.Error("invalid slave %s, lastlogid %d > %d", s.slaveListeningAddr, s.lastLogID, logId)
		}
	}

	app.slock.Unlock()

	if n >= total {
		//at least total slaves have owned this log
		return
	}

	startTime := time.Now()
	done := make(chan struct{}, 1)
	go func() {
		n := 0
		for i := 0; i < slaveNum; i++ {
			id := <-app.slaveSyncAck
			if id < logId {
				log.Info("some slave may close with last logid %d < %d", id, logId)
			} else {
				n++
				if n >= total {
					break
				}
			}
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(time.Duration(app.cfg.Replication.WaitSyncTime) * time.Millisecond):
		log.Info("replication wait timeout")
	}

	stopTime := time.Now()
	app.info.Replication.PubLogAckNum.Add(1)
	app.info.Replication.PubLogTotalAckTime.Add(stopTime.Sub(startTime))
}
