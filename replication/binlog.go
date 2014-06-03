package replication

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-log/log"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	MaxBinLogFileSize int = 1024 * 1024 * 1024
	MaxBinLogFileNum  int = 10000

	DefaultBinLogFileSize int = MaxBinLogFileSize
	DefaultBinLogFileNum  int = 10
)

/*
index file format:
ledis-bin.00001
ledis-bin.00002
ledis-bin.00003

log file format

timestamp(bigendian uint32, seconds)|PayloadLen(bigendian uint32)|PayloadData|LogId

*/

type BinLogConfig struct {
	BaseName    string `json:"base_name"`
	IndexName   string `json:"index_name"`
	Path        string `json:"path"`
	MaxFileSize int    `json:"max_file_size"`
	MaxFileNum  int    `json:"max_file_num"`
}

func (cfg *BinLogConfig) adjust() {
	if cfg.MaxFileSize <= 0 {
		cfg.MaxFileSize = DefaultBinLogFileSize
	} else if cfg.MaxFileSize > MaxBinLogFileSize {
		cfg.MaxFileSize = MaxBinLogFileSize
	}

	if cfg.MaxFileNum <= 0 {
		cfg.MaxFileNum = DefaultBinLogFileNum
	} else if cfg.MaxFileNum > MaxBinLogFileNum {
		cfg.MaxFileNum = MaxBinLogFileNum
	}

	if len(cfg.BaseName) == 0 {
		cfg.BaseName = "ledis"
	}
	if len(cfg.IndexName) == 0 {
		cfg.IndexName = "ledis"
	}
}

type BinLog struct {
	cfg *BinLogConfig

	logFile *os.File

	logWb *bufio.Writer

	indexName    string
	logNames     []string
	lastLogIndex int
}

func NewBinLog(data json.RawMessage) (*BinLog, error) {
	var cfg BinLogConfig

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return NewBinLogWithConfig(&cfg)
}

func NewBinLogWithConfig(cfg *BinLogConfig) (*BinLog, error) {
	b := new(BinLog)

	cfg.adjust()

	b.cfg = cfg

	if err := os.MkdirAll(cfg.Path, os.ModePerm); err != nil {
		return nil, err
	}

	b.logNames = make([]string, 0, b.cfg.MaxFileNum)

	if err := b.loadIndex(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *BinLog) Close() {
	if b.logFile != nil {
		b.logFile.Close()
	}
}

func (b *BinLog) deleteOldest() {
	logPath := path.Join(b.cfg.Path, b.logNames[0])
	os.Remove(logPath)

	copy(b.logNames[0:], b.logNames[1:])
	b.logNames = b.logNames[0 : len(b.logNames)-1]
}

func (b *BinLog) flushIndex() error {
	data := strings.Join(b.logNames, "\n")

	bakName := fmt.Sprintf("%s.bak", b.indexName)
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

	if err := os.Rename(bakName, b.indexName); err != nil {
		log.Error("rename binlog bak index error %s", err.Error())
		return err
	}

	return nil
}

func (b *BinLog) loadIndex() error {
	b.indexName = path.Join(b.cfg.Path, fmt.Sprintf("%s-bin.index", b.cfg.IndexName))
	fd, err := os.OpenFile(b.indexName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	//maybe we will check valid later?
	rb := bufio.NewReader(fd)
	for {
		line, err := rb.ReadString('\n')
		if err != nil && err != io.EOF {
			fd.Close()
			return err
		}

		line = strings.Trim(line, "\r\n ")

		if len(line) > 0 {
			b.logNames = append(b.logNames, line)
		}

		if len(b.logNames) == b.cfg.MaxFileNum {
			//remove oldest logfile
			b.deleteOldest()
		}

		if err == io.EOF {
			break
		}
	}

	fd.Close()

	if err := b.flushIndex(); err != nil {
		return err
	}

	if len(b.logNames) == 0 {
		b.lastLogIndex = 1
	} else {
		lastName := b.logNames[len(b.logNames)-1]

		if b.lastLogIndex, err = strconv.Atoi(path.Ext(lastName)[1:]); err != nil {
			log.Error("invalid logfile name %s", err.Error())
			return err
		}

		//like mysql, if server restart, a new binlog will create
		b.lastLogIndex++
	}

	return nil
}

func (b *BinLog) getLogName() string {
	return fmt.Sprintf("%s-bin.%05d", b.cfg.BaseName, b.lastLogIndex)
}

func (b *BinLog) openNewLogFile() error {
	var err error
	lastName := b.getLogName()

	logPath := path.Join(b.cfg.Path, lastName)
	if b.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
		log.Error("open new logfile error %s", err.Error())
		return err
	}

	if len(b.logNames) == b.cfg.MaxFileNum {
		b.deleteOldest()
	}

	b.logNames = append(b.logNames, lastName)

	if b.logWb == nil {
		b.logWb = bufio.NewWriterSize(b.logFile, 1024)
	} else {
		b.logWb.Reset(b.logFile)
	}

	if err = b.flushIndex(); err != nil {
		return err
	}

	return nil
}

func (b *BinLog) openLogFile() error {
	if b.logFile == nil {
		return b.openNewLogFile()
	} else {
		//check file size
		st, _ := b.logFile.Stat()
		if st.Size() >= int64(b.cfg.MaxFileSize) {
			//must use new file
			b.lastLogIndex++

			b.logFile.Close()

			return b.openNewLogFile()
		}
	}

	return nil
}

func (b *BinLog) Log(args ...[]byte) error {
	var err error

	if err = b.openLogFile(); err != nil {
		return err
	}

	for _, data := range args {
		createTime := uint32(time.Now().Unix())
		payLoadLen := len(data)

		binary.Write(b.logWb, binary.BigEndian, createTime)
		binary.Write(b.logWb, binary.BigEndian, payLoadLen)

		b.logWb.Write(data)

		if err = b.logWb.Flush(); err != nil {
			log.Error("write log error %s", err.Error())
			return err
		}
	}

	return nil
}

func (b *BinLog) LogFileName() string {
	if len(b.logNames) == 0 {
		return ""
	} else {
		return b.logNames[len(b.logNames)-1]
	}
}

func (b *BinLog) LogFilePos() int64 {
	if b.logFile == nil {
		return 0
	} else {
		st, _ := b.logFile.Stat()
		return st.Size()
	}
}
