package server

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/go-snappy/snappy"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/rpl"
	"net"
	"os"
	"path"
	"strconv"
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

	compressBuf []byte
}

func newMaster(app *App) *master {
	m := new(master)
	m.app = app

	m.quit = make(chan struct{}, 1)

	m.compressBuf = make([]byte, 256)

	return m
}

func (m *master) Close() {
	select {
	case m.quit <- struct{}{}:
	default:
	}

	if m.conn != nil {
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

	m.app.ldb.SetReadOnly(false)

	return nil
}

func (m *master) startReplication(masterAddr string) error {
	//stop last replcation, if avaliable
	m.Close()

	m.addr = masterAddr

	m.quit = make(chan struct{}, 1)

	m.app.ldb.SetReadOnly(true)

	go m.runReplication()
	return nil
}

func (m *master) runReplication() {
	m.wg.Add(1)
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

		for {
			if err := m.sync(); err != nil {
				if m.conn != nil {
					//if conn == nil, other close the replication, not error
					log.Warn("sync error %s", err.Error())
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
	fullSyncCmd   = []byte("*1\r\n$8\r\nfullsync\r\n")  //fullsync
	syncCmdFormat = "*2\r\n$4\r\nsync\r\n$%d\r\n%s\r\n" //sync logid
)

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

func (m *master) sync() error {
	var err error
	var syncID uint64
	if syncID, err = m.app.ldb.NextSyncLogID(); err != nil {
		return err
	}

	logIDStr := strconv.FormatUint(syncID, 10)

	cmd := ledis.Slice(fmt.Sprintf(syncCmdFormat, len(logIDStr),
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

	var buf []byte
	buf, err = snappy.Decode(m.compressBuf, m.syncBuf.Bytes())
	if err != nil {
		return err
	} else if len(buf) > len(m.compressBuf) {
		m.compressBuf = buf
	}

	if len(buf) == 0 {
		return nil
	}

	if err = m.app.ldb.StoreLogsFromData(buf); err != nil {
		return err
	}

	return nil

}

func (app *App) slaveof(masterAddr string) error {
	app.m.Lock()
	defer app.m.Unlock()

	if !app.ldb.ReplicationUsed() {
		return fmt.Errorf("slaveof must enable replication")
	}

	if len(masterAddr) == 0 {
		return app.m.stopReplication()
	} else {
		return app.m.startReplication(masterAddr)
	}

	return nil
}

func (app *App) addSlave(c *client) {
	app.slock.Lock()
	defer app.slock.Unlock()

	app.slaves[c] = struct{}{}
}

func (app *App) removeSlave(c *client) {
	app.slock.Lock()
	defer app.slock.Unlock()

	delete(app.slaves, c)

	if c.ack != nil {
		select {
		case c.ack.ch <- c.lastLogID:
		default:
		}
	}
}

func (app *App) publishNewLog(l *rpl.Log) {
	if !app.cfg.Replication.Sync {
		//no sync replication, we will do async
		return
	}

	ss := make([]*client, 0, 4)
	app.slock.Lock()

	logId := l.ID
	for s, _ := range app.slaves {
		if s.lastLogID >= logId {
			//slave has already this log
			ss = []*client{}
			break
		} else {
			ss = append(ss, s)
		}
	}

	app.slock.Unlock()

	if len(ss) == 0 {
		return
	}

	ack := &syncAck{
		logId, make(chan uint64, len(ss)),
	}

	for _, s := range ss {
		s.ack = ack
	}

	done := make(chan struct{}, 1)
	go func() {
		for i := 0; i < len(ss); i++ {
			id := <-ack.ch
			if id > logId {
				break
			}
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(time.Duration(app.cfg.Replication.WaitSyncTime) * time.Second):
	}
}
