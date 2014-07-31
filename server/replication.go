package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/go-snappy/snappy"
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

type MasterInfo struct {
	Addr         string `json:"addr"`
	LogFileIndex int64  `json:"log_file_index"`
	LogPos       int64  `json:"log_pos"`
}

func (m *MasterInfo) Save(filePath string) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	filePathBak := fmt.Sprintf("%s.bak", filePath)

	var fd *os.File
	fd, err = os.OpenFile(filePathBak, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	if _, err = fd.Write(data); err != nil {
		fd.Close()
		return err
	}

	fd.Close()
	return os.Rename(filePathBak, filePath)
}

func (m *MasterInfo) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
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

type master struct {
	sync.Mutex

	conn net.Conn
	rb   *bufio.Reader

	app *App

	quit chan struct{}

	infoName string

	info *MasterInfo

	wg sync.WaitGroup

	syncBuf bytes.Buffer

	compressBuf []byte
}

func newMaster(app *App) *master {
	m := new(master)
	m.app = app

	m.infoName = path.Join(m.app.cfg.DataDir, "master.info")

	m.quit = make(chan struct{}, 1)

	m.compressBuf = make([]byte, 256)

	m.info = new(MasterInfo)

	//if load error, we will start a fullsync later
	m.loadInfo()

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

func (m *master) loadInfo() error {
	return m.info.Load(m.infoName)
}

func (m *master) saveInfo() error {
	return m.info.Save(m.infoName)
}

func (m *master) connect() error {
	if len(m.info.Addr) == 0 {
		return fmt.Errorf("no assign master addr")
	}

	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}

	if conn, err := net.Dial("tcp", m.info.Addr); err != nil {
		return err
	} else {
		m.conn = conn

		m.rb = bufio.NewReaderSize(m.conn, 4096)
	}
	return nil
}

func (m *master) resetInfo(addr string) {
	m.info.Addr = addr
	m.info.LogFileIndex = 0
	m.info.LogPos = 0
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

	if masterAddr != m.info.Addr {
		m.resetInfo(masterAddr)
		if err := m.saveInfo(); err != nil {
			log.Error("save master info error %s", err.Error())
			return err
		}
	}

	m.quit = make(chan struct{}, 1)

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
				log.Error("connect master %s error %s, try 2s later", m.info.Addr, err.Error())
				time.Sleep(2 * time.Second)
				continue
			}
		}

		if m.info.LogFileIndex == 0 {
			//try a fullsync
			if err := m.fullSync(); err != nil {
				log.Warn("full sync error %s", err.Error())
				return
			}

			if m.info.LogFileIndex == 0 {
				//master not support binlog, we cannot sync, so stop replication
				m.stopReplication()
				return
			}
		}

		for {
			for {
				lastIndex := m.info.LogFileIndex
				lastPos := m.info.LogPos
				if err := m.sync(); err != nil {
					log.Warn("sync error %s", err.Error())
					return
				}

				if m.info.LogFileIndex == lastIndex && m.info.LogPos == lastPos {
					//sync no data, wait 1s and retry
					break
				}
			}

			select {
			case <-m.quit:
				return
			case <-time.After(1 * time.Second):
				break
			}
		}
	}

	return
}

var (
	fullSyncCmd   = []byte("*1\r\n$8\r\nfullsync\r\n")               //fullsync
	syncCmdFormat = "*3\r\n$4\r\nsync\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n" //sync index pos
)

func (m *master) fullSync() error {
	if _, err := m.conn.Write(fullSyncCmd); err != nil {
		return err
	}

	dumpPath := path.Join(m.app.cfg.DataDir, "master.dump")
	f, err := os.OpenFile(dumpPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
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

	if err = m.app.ldb.FlushAll(); err != nil {
		return err
	}

	var head *ledis.MasterInfo
	head, err = m.app.ldb.LoadDumpFile(dumpPath)

	if err != nil {
		log.Error("load dump file error %s", err.Error())
		return err
	}

	m.info.LogFileIndex = head.LogFileIndex
	m.info.LogPos = head.LogPos

	return m.saveInfo()
}

func (m *master) sync() error {
	logIndexStr := strconv.FormatInt(m.info.LogFileIndex, 10)
	logPosStr := strconv.FormatInt(m.info.LogPos, 10)

	cmd := ledis.Slice(fmt.Sprintf(syncCmdFormat, len(logIndexStr),
		logIndexStr, len(logPosStr), logPosStr))
	if _, err := m.conn.Write(cmd); err != nil {
		return err
	}

	m.syncBuf.Reset()

	err := ReadBulkTo(m.rb, &m.syncBuf)
	if err != nil {
		return err
	}

	var buf []byte
	buf, err = snappy.Decode(m.compressBuf, m.syncBuf.Bytes())
	if err != nil {
		return err
	} else if len(buf) > len(m.compressBuf) {
		m.compressBuf = buf
	}

	if len(buf) < 16 {
		return fmt.Errorf("invalid sync data len %d", len(buf))
	}

	m.info.LogFileIndex = int64(binary.BigEndian.Uint64(buf[0:8]))
	m.info.LogPos = int64(binary.BigEndian.Uint64(buf[8:16]))

	if m.info.LogFileIndex == 0 {
		//master now not support binlog, stop replication
		m.stopReplication()
		return nil
	} else if m.info.LogFileIndex == -1 {
		//-1 means than binlog index and pos are lost, we must start a full sync instead
		return m.fullSync()
	}

	err = m.app.ldb.ReplicateFromData(buf[16:])
	if err != nil {
		return err
	}

	return m.saveInfo()

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
