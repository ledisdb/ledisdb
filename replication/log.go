package replication

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/siddontang/go-log/log"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	ErrOverSpaceLimit = errors.New("total log files exceed space limit")
)

type logHandler interface {
	Write(wb *bufio.Writer, data []byte) (int, error)
}

type LogConfig struct {
	BaseName    string `json:"base_name"`
	IndexName   string `json:"index_name"`
	LogType     string `json:"log_type"`
	Path        string `json:"path"`
	MaxFileSize int    `json:"max_file_size"`
	MaxFileNum  int    `json:"max_file_num"`
	SpaceLimit  int64  `json:"space_limit"`
}

type Log struct {
	cfg *LogConfig

	logFile *os.File

	logWb *bufio.Writer

	indexName    string
	logNames     []string
	lastLogIndex int

	space int64

	handler logHandler
}

func newLog(handler logHandler, cfg *LogConfig) (*Log, error) {
	l := new(Log)

	l.cfg = cfg
	l.handler = handler

	if err := os.MkdirAll(cfg.Path, os.ModePerm); err != nil {
		return nil, err
	}

	l.logNames = make([]string, 0, 16)
	l.space = 0

	if err := l.loadIndex(); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *Log) flushIndex() error {
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

func (l *Log) loadIndex() error {
	l.indexName = path.Join(l.cfg.Path, fmt.Sprintf("%s-%s.index", l.cfg.IndexName, l.cfg.LogType))
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

			if st, err := os.Stat(path.Join(l.cfg.Path, line)); err != nil {
				log.Error("load index line %s error %s", line, err.Error())
				return err
			} else {
				l.space += st.Size()

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

		if l.lastLogIndex, err = strconv.Atoi(path.Ext(lastName)[1:]); err != nil {
			log.Error("invalid logfile name %s", err.Error())
			return err
		}

		//like mysql, if server restart, a new binlog will create
		l.lastLogIndex++
	}

	return nil
}

func (l *Log) getLogFile() string {
	return fmt.Sprintf("%s-%s.%07d", l.cfg.BaseName, l.cfg.LogType, l.lastLogIndex)
}

func (l *Log) openNewLogFile() error {
	var err error
	lastName := l.getLogFile()

	logPath := path.Join(l.cfg.Path, lastName)
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

func (l *Log) checkLogFileSize() bool {
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

func (l *Log) purge(n int) {
	for i := 0; i < n; i++ {
		logPath := path.Join(l.cfg.Path, l.logNames[i])
		if st, err := os.Stat(logPath); err != nil {
			log.Error("purge %s error %s", logPath, err.Error())
		} else {
			l.space -= st.Size()
		}

		os.Remove(logPath)
	}

	copy(l.logNames[0:], l.logNames[n:])
	l.logNames = l.logNames[0 : len(l.logNames)-n]
}

func (l *Log) Close() {
	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
	}
}

func (l *Log) LogNames() []string {
	return l.logNames
}

func (l *Log) LogFileName() string {
	return l.getLogFile()
}

func (l *Log) LogFilePos() int64 {
	if l.logFile == nil {
		return 0
	} else {
		st, _ := l.logFile.Stat()
		return st.Size()
	}
}

func (l *Log) Purge(n int) error {
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

func (l *Log) Log(args ...[]byte) error {
	if l.cfg.SpaceLimit > 0 && l.space >= l.cfg.SpaceLimit {
		return ErrOverSpaceLimit
	}

	var err error

	if l.logFile == nil {
		if err = l.openNewLogFile(); err != nil {
			return err
		}
	}

	totalSize := 0

	var n int = 0
	for _, data := range args {
		if n, err = l.handler.Write(l.logWb, data); err != nil {
			log.Error("write log error %s", err.Error())
			return err
		} else {
			totalSize += n
		}

	}

	if err = l.logWb.Flush(); err != nil {
		log.Error("write log error %s", err.Error())
		return err
	}

	l.space += int64(totalSize)

	l.checkLogFileSize()

	return nil
}
