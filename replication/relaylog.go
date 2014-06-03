package replication

import (
	"bufio"
	"encoding/json"
)

const (
	MaxRelayLogFileSize     int = 1024 * 1024 * 1024
	DefaultRelayLogFileSize int = MaxRelayLogFileSize
)

type RelayLogConfig struct {
	LogConfig
}

func (cfg *RelayLogConfig) adjust() {
	if cfg.MaxFileSize <= 0 {
		cfg.MaxFileSize = DefaultRelayLogFileSize
	} else if cfg.MaxFileSize > MaxRelayLogFileSize {
		cfg.MaxFileSize = MaxRelayLogFileSize
	}

	if len(cfg.BaseName) == 0 {
		cfg.BaseName = "ledis"
	}
	if len(cfg.IndexName) == 0 {
		cfg.IndexName = "ledis"
	}

	cfg.MaxFileNum = -1
	cfg.LogType = "relay"
}

type relayLogHandler struct {
}

func (h *relayLogHandler) Write(wb *bufio.Writer, data []byte) (int, error) {
	return wb.Write(data)
}

func NewRelayLog(data json.RawMessage) (*Log, error) {
	var cfg RelayLogConfig

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return NewRelayLogWithConfig(&cfg)
}

func NewRelayLogWithConfig(cfg *RelayLogConfig) (*Log, error) {
	cfg.adjust()

	return newLog(new(relayLogHandler), &cfg.LogConfig)
}
