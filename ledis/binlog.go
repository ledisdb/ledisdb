package ledis

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/config"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

/*
index file format:
ledis-bin.00001
ledis-bin.00002
ledis-bin.00003

log file format

timestamp(bigendian uint32, seconds)|PayloadLen(bigendian uint32)|PayloadData

*/

type BinLog struct {
	path string

	cfg *config.BinLogConfig

	logFile *os.File

	logWb *bufio.Writer

	indexName    string
	logNames     []string
	lastLogIndex int64
}

func NewBinLog(cfg *config.Config) (*BinLog, error) {
	l := new(BinLog)

	l.cfg = &cfg.BinLog
	l.cfg.Adjust()

	l.path = path.Join(cfg.DataDir, "bin_log")

	if err := os.MkdirAll(l.path, os.ModePerm); err != nil {
		return nil, err
	}

	l.logNames = make([]string, 0, 16)

	if err := l.loadIndex(); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *BinLog) flushIndex() error {
	data := strings.Join(l.logNames, "\n")

	bakName := fmt.Sprintf("%s.bak", l.indexName)
	f, err := os.OpenFile(bakName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error("create binlog bak index error %s", err.Error())
		return err
	}

	if _, err := f.WriteString(data); err != nil {
		log.Error("write binlog index error %s", err.Error())
		f.Close()
		return err
	}

	f.Close()

	if err := os.Rename(bakName, l.indexName); err != nil {
		log.Error("rename binlog bak index error %s", err.Error())
		return err
	}

	return nil
}

func (l *BinLog) loadIndex() error {
	l.indexName = path.Join(l.path, fmt.Sprintf("ledis-bin.index"))
	if _, err := os.Stat(l.indexName); os.IsNotExist(err) {
		//no index file, nothing to do
	} else {
		indexData, err := ioutil.ReadFile(l.indexName)
		if err != nil {
			return err
		}

		lines := strings.Split(string(indexData), "\n")
		for _, line := range lines {
			line = strings.Trim(line, "\r\n ")
			if len(line) == 0 {
				continue
			}

			if _, err := os.Stat(path.Join(l.path, line)); err != nil {
				log.Error("load index line %s error %s", line, err.Error())
				return err
			} else {
				l.logNames = append(l.logNames, line)
			}
		}
	}
	if l.cfg.MaxFileNum > 0 && len(l.logNames) > l.cfg.MaxFileNum {
		//remove oldest logfile
		if err := l.Purge(len(l.logNames) - l.cfg.MaxFileNum); err != nil {
			return err
		}
	}

	var err error
	if len(l.logNames) == 0 {
		l.lastLogIndex = 1
	} else {
		lastName := l.logNames[len(l.logNames)-1]

		if l.lastLogIndex, err = strconv.ParseInt(path.Ext(lastName)[1:], 10, 64); err != nil {
			log.Error("invalid logfile name %s", err.Error())
			return err
		}

		//like mysql, if server restart, a new binlog will create
		l.lastLogIndex++
	}

	return nil
}

func (l *BinLog) getLogFile() string {
	return l.FormatLogFileName(l.lastLogIndex)
}

func (l *BinLog) openNewLogFile() error {
	var err error
	lastName := l.getLogFile()

	logPath := path.Join(l.path, lastName)
	if l.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
		log.Error("open new logfile error %s", err.Error())
		return err
	}

	if l.cfg.MaxFileNum > 0 && len(l.logNames) == l.cfg.MaxFileNum {
		l.purge(1)
	}

	l.logNames = append(l.logNames, lastName)

	if l.logWb == nil {
		l.logWb = bufio.NewWriterSize(l.logFile, 1024)
	} else {
		l.logWb.Reset(l.logFile)
	}

	if err = l.flushIndex(); err != nil {
		return err
	}

	return nil
}

func (l *BinLog) checkLogFileSize() bool {
	if l.logFile == nil {
		return false
	}

	st, _ := l.logFile.Stat()
	if st.Size() >= int64(l.cfg.MaxFileSize) {
		l.lastLogIndex++

		l.logFile.Close()
		l.logFile = nil
		return true
	}

	return false
}

func (l *BinLog) purge(n int) {
	for i := 0; i < n; i++ {
		logPath := path.Join(l.path, l.logNames[i])
		os.Remove(logPath)
	}

	copy(l.logNames[0:], l.logNames[n:])
	l.logNames = l.logNames[0 : len(l.logNames)-n]
}

func (l *BinLog) Close() {
	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
	}
}

func (l *BinLog) LogNames() []string {
	return l.logNames
}

func (l *BinLog) LogFileName() string {
	return l.getLogFile()
}

func (l *BinLog) LogFilePos() int64 {
	if l.logFile == nil {
		return 0
	} else {
		st, _ := l.logFile.Stat()
		return st.Size()
	}
}

func (l *BinLog) LogFileIndex() int64 {
	return l.lastLogIndex
}

func (l *BinLog) FormatLogFileName(index int64) string {
	return fmt.Sprintf("ledis-bin.%07d", index)
}

func (l *BinLog) FormatLogFilePath(index int64) string {
	return path.Join(l.path, l.FormatLogFileName(index))
}

func (l *BinLog) LogPath() string {
	return l.path
}

func (l *BinLog) Purge(n int) error {
	if len(l.logNames) == 0 {
		return nil
	}

	if n >= len(l.logNames) {
		n = len(l.logNames)
		//can not purge current log file
		if l.logNames[n-1] == l.getLogFile() {
			n = n - 1
		}
	}

	l.purge(n)

	return l.flushIndex()
}

func (l *BinLog) Log(args ...[]byte) error {
	var err error

	if l.logFile == nil {
		if err = l.openNewLogFile(); err != nil {
			return err
		}
	}

	//we treat log many args as a batch, so use same createTime
	createTime := uint32(time.Now().Unix())

	for _, data := range args {
		payLoadLen := uint32(len(data))

		if err := binary.Write(l.logWb, binary.BigEndian, createTime); err != nil {
			return err
		}

		if err := binary.Write(l.logWb, binary.BigEndian, payLoadLen); err != nil {
			return err
		}

		if _, err := l.logWb.Write(data); err != nil {
			return err
		}
	}

	if err = l.logWb.Flush(); err != nil {
		log.Error("write log error %s", err.Error())
		return err
	}

	l.checkLogFileSize()

	return nil
}
