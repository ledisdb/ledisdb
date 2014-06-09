package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/ledis"
	"io/ioutil"
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

	addr         string `json:"addr"`
	logFileIndex int64  `json:"log_file_index"`
	logPos       int64  `json:"log_pos"`

	c  net.Conn
	rb *bufio.Reader

	app *App

	quit chan struct{}

	infoName    string
	infoNameBak string

	wg sync.WaitGroup

	syncBuf bytes.Buffer
}

func newMaster(app *App) *master {
	m := new(master)
	m.app = app

	m.infoName = path.Join(m.app.cfg.DataDir, "master.info")
	m.infoNameBak = fmt.Sprintf("%s.bak", m.infoName)

	m.quit = make(chan struct{})

	//if load error, we will start a fullsync later
	m.loadInfo()

	return m
}

func (m *master) Close() {
	close(m.quit)

	if m.c != nil {
		m.c.Close()
		m.c = nil
	}

	m.wg.Wait()
}

func (m *master) loadInfo() error {
	data, err := ioutil.ReadFile(m.infoName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	if err = json.Unmarshal(data, m); err != nil {
		return err
	}

	return nil
}

func (m *master) saveInfo() error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	var fd *os.File
	fd, err = os.OpenFile(m.infoNameBak, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	if _, err = fd.Write(data); err != nil {
		fd.Close()
		return err
	}

	fd.Close()
	return os.Rename(m.infoNameBak, m.infoName)
}

func (m *master) connect() error {
	if len(m.addr) == 0 {
		return fmt.Errorf("no assign master addr")
	}

	if m.c != nil {
		m.c.Close()
		m.c = nil
	}

	if c, err := net.Dial("tcp", m.addr); err != nil {
		return err
	} else {
		m.c = c

		m.rb = bufio.NewReaderSize(m.c, 4096)
	}
	return nil
}

func (m *master) resetInfo(addr string) {
	m.addr = addr
	m.logFileIndex = 0
	m.logPos = 0
}

func (m *master) stopReplication() error {
	m.Close()

	if err := m.saveInfo(); err != nil {
		log.Error("save master info error %s", err.Error())
		return err
	}

	return nil
}

func (m *master) startReplication(masterAddr string) error {
	//stop last replcation, if avaliable
	m.Close()

	if masterAddr != m.addr {
		m.resetInfo(masterAddr)
		if err := m.saveInfo(); err != nil {
			log.Error("save master info error %s", err.Error())
			return err
		}
	}

	m.quit = make(chan struct{})

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

		if m.logFileIndex == 0 {
			//try a fullsync
			if err := m.fullSync(); err != nil {
				log.Warn("full sync error %s", err.Error())
				return
			}

			if m.logFileIndex == 0 {
				//master not support binlog, we cannot sync, so stop replication
				m.stopReplication()
				return
			}
		}

		t := time.NewTicker(1 * time.Second)

		//then we will try sync every 1 seconds
		for {
			select {
			case <-t.C:
				if err := m.sync(); err != nil {
					log.Warn("sync error %s", err.Error())
					return
				}
			case <-m.quit:
				return
			}
		}
	}

	return
}

var (
	fullSyncCmd   = []byte("*1\r\n$8\r\nfullsync\r\n")              //fullsync
	syncCmdFormat = "*3\r\n$4\r\nsync\r\n$%d\r\n%s\r\n%d\r\n%s\r\n" //sync file pos
)

func (m *master) fullSync() error {
	if _, err := m.c.Write(fullSyncCmd); err != nil {
		return err
	}

	dumpPath := path.Join(m.app.cfg.DataDir, "master.dump")
	f, err := os.OpenFile(dumpPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer os.Remove(dumpPath)

	err = readBulkTo(m.rb, f)
	f.Close()
	if err != nil {
		log.Error("read dump data error %s", err.Error())
		return err
	}

	if err = m.app.ldb.FlushAll(); err != nil {
		return err
	}

	var head *ledis.MasterInfo
	head, err = m.app.ldb.LoadDumpFile(dumpPath)

	if err != nil {
		log.Error("load dump file error %s", err.Error())
		return err
	}

	m.logFileIndex = head.LogFileIndex
	m.logPos = head.LogPos

	return nil
}

func (m *master) sync() error {
	logIndexStr := strconv.FormatInt(m.logFileIndex, 10)
	logPosStr := strconv.FormatInt(m.logPos, 10)

	if _, err := m.c.Write(ledis.Slice(fmt.Sprintf(syncCmdFormat, len(logIndexStr),
		logIndexStr, len(logPosStr), logPosStr))); err != nil {
		return err
	}

	m.syncBuf.Reset()

	err := readBulkTo(m.rb, &m.syncBuf)
	if err != nil {
		return err
	}

	err = binary.Read(&m.syncBuf, binary.BigEndian, &m.logFileIndex)
	if err != nil {
		return err
	}

	err = binary.Read(&m.syncBuf, binary.BigEndian, &m.logPos)
	if err != nil {
		return err
	}

	if m.logFileIndex == 0 {
		//master now not support binlog, stop replication
		m.stopReplication()
		return nil
	} else if m.logFileIndex == -1 {
		//-1 means than binlog index and pos are lost, we must start a full sync instead
		return m.fullSync()
	}

	err = m.app.ldb.ReplicateFromReader(&m.syncBuf)
	if err != nil {
		return err
	}

	return nil

}

func (app *App) slaveof(masterAddr string) error {
	app.m.Lock()
	defer app.m.Unlock()

	if len(masterAddr) == 0 {
		return app.m.stopReplication()
	} else {
		return app.m.startReplication(masterAddr)
	}

	return nil
}
