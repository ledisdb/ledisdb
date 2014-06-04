package replication

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
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

timestamp(bigendian uint32, seconds)|PayloadLen(bigendian uint32)|PayloadData

*/

type BinLogConfig struct {
	LogConfig
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

	//binlog not care space limit
	cfg.SpaceLimit = -1

	cfg.LogType = "bin"
}

type binlogHandler struct {
}

func (h *binlogHandler) Write(wb *bufio.Writer, data []byte) (int, error) {
	createTime := uint32(time.Now().Unix())
	payLoadLen := uint32(len(data))

	if err := binary.Write(wb, binary.BigEndian, createTime); err != nil {
		return 0, err
	}

	if err := binary.Write(wb, binary.BigEndian, payLoadLen); err != nil {
		return 0, err
	}

	if _, err := wb.Write(data); err != nil {
		return 0, err
	}

	return 8 + len(data), nil
}

func NewBinLog(data json.RawMessage) (*Log, error) {
	var cfg BinLogConfig

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return NewBinLogWithConfig(&cfg)
}

func NewBinLogWithConfig(cfg *BinLogConfig) (*Log, error) {
	cfg.adjust()

	return newLog(new(binlogHandler), &cfg.LogConfig)
}
