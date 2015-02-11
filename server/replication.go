package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/siddontang/go/log"
	"github.com/siddontang/go/num"
	"github.com/siddontang/go/sync2"
	goledis "github.com/siddontang/ledisdb/client/go/ledis"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/rpl"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var (
	errConnectMaster = errors.New("connect master error")
	errReplClosed    = errors.New("replication is closed")
)

const (
	// slave needs to connect to its master
	replConnectState int32 = iota + 1
	// slave-master connection is in progress
	replConnectingState
	// perform the synchronization
	replSyncState
	// slave is online
	replConnectedState
)

type syncBuffer struct {
	m *master
	bytes.Buffer
}

func (b *syncBuffer) Write(data []byte) (int, error) {
	b.m.state.Set(replSyncState)
	n, err := b.Buffer.Write(data)
	return n, err
}

type master struct {
	sync.Mutex

	conn *goledis.Conn

	app *App

	quit chan struct{}

	addr string

	wg sync.WaitGroup

	syncBuf syncBuffer

	state sync2.AtomicInt32
}

func newMaster(app *App) *master {
	m := new(master)
	m.app = app

	m.quit = make(chan struct{}, 1)
	m.syncBuf = syncBuffer{m: m}

	m.state.Set(replConnectState)

	return m
}

func (m *master) Close() {
	m.quit <- struct{}{}

	m.closeConn()

	m.wg.Wait()

	select {
	case <-m.quit:
	default:
	}

	m.state.Set(replConnectState)
}

func (m *master) resetConn() error {
	if len(m.addr) == 0 {
		return fmt.Errorf("no assign master addr")
	}

	if m.conn != nil {
		m.conn.Close()
	}

	m.conn = goledis.NewConn(m.addr)

	return nil
}

func (m *master) closeConn() {
	if m.conn != nil {
		//for replication, we send quit command to close gracefully
		m.conn.Send("quit")

		m.conn.Close()
	}
}

func (m *master) stopReplication() error {
	m.Close()

	return nil
}

func (m *master) startReplication(masterAddr string, restart bool) error {
	//stop last replcation, if avaliable
	m.Close()

	m.addr = masterAddr

	m.app.cfg.SetReadonly(true)

	m.wg.Add(1)
	go m.runReplication(restart)
	return nil
}

func (m *master) needQuit() bool {
	select {
	case <-m.quit:
		return true
	default:
		return false
	}
}

func (m *master) runReplication(restart bool) {
	defer func() {
		m.state.Set(replConnectState)
		m.wg.Done()
	}()

	if err := m.resetConn(); err != nil {
		log.Errorf("reset conn error %s", err.Error())
		return
	}

	for {
		m.state.Set(replConnectState)

		if _, err := m.conn.Do("ping"); err != nil {
			log.Errorf("ping master %s error %s, try 3s later", m.addr, err.Error())
			time.Sleep(3 * time.Second)
			continue
		}

		if m.needQuit() {
			return
		}

		m.state.Set(replConnectedState)

		if err := m.replConf(); err != nil {
			log.Errorf("replconf error %s", err.Error())
			continue
		}

		if restart {
			if err := m.fullSync(); err != nil {
				log.Errorf("restart fullsync error %s", err.Error())
				continue
			}
			m.state.Set(replConnectedState)
		}

		for {
			if err := m.sync(); err != nil {
				log.Errorf("sync error %s", err.Error())
				break
			}
			m.state.Set(replConnectedState)

			if m.needQuit() {
				return
			}
		}
	}

	return
}

func (m *master) replConf() error {
	_, port, err := net.SplitHostPort(m.app.cfg.Addr)
	if err != nil {
		return err
	}

	if s, err := goledis.String(m.conn.Do("replconf", "listening-port", port)); err != nil {
		return err
	} else if strings.ToUpper(s) != "OK" {
		return fmt.Errorf("not ok but %s", s)
	}

	return nil
}

func (m *master) fullSync() error {
	log.Info("begin full sync")

	if err := m.conn.Send("fullsync"); err != nil {
		return err
	}

	m.state.Set(replSyncState)

	dumpPath := path.Join(m.app.cfg.DataDir, "master.dump")
	f, err := os.OpenFile(dumpPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer os.Remove(dumpPath)

	err = m.conn.ReceiveBulkTo(f)
	f.Close()
	if err != nil {
		log.Errorf("read dump data error %s", err.Error())
		return err
	}

	if _, err = m.app.ldb.LoadDumpFile(dumpPath); err != nil {
		log.Errorf("load dump file error %s", err.Error())
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

	if err := m.conn.Send("sync", syncID); err != nil {
		return err
	}

	m.state.Set(replConnectedState)

	m.syncBuf.Reset()

	if err = m.conn.ReceiveBulkTo(&m.syncBuf); err != nil {
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

	m.state.Set(replConnectedState)

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
		app.cfg.SetReadonly(readonly)
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

		app.cfg.SetReadonly(readonly)
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
		log.Infof("remove slave %s", addr)
		if activeQuit {
			asyncNotifyUint64(app.slaveSyncAck, c.lastLogID.Get())
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

	asyncNotifyUint64(app.slaveSyncAck, c.lastLogID.Get())
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
		lastLogID := s.lastLogID.Get()
		if lastLogID == logId {
			//slave has already owned this log
			n++
		} else if lastLogID > logId {
			log.Errorf("invalid slave %s, lastlogid %d > %d", s.slaveListeningAddr, lastLogID, logId)
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
				log.Infof("some slave may close with last logid %d < %d", id, logId)
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
