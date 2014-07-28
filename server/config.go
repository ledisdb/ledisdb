package server

import (
	"encoding/json"
	"github.com/siddontang/copier"
	"github.com/siddontang/ledisdb/ledis"
	"io/ioutil"
)

type Config struct {
	Addr string `json:"addr"`

	DataDir string `json:"data_dir"`

	DB struct {
		Name            string `json:"name"`
		Compression     bool   `json:"compression"`
		BlockSize       int    `json:"block_size"`
		WriteBufferSize int    `json:"write_buffer_size"`
		CacheSize       int    `json:"cache_size"`
		MaxOpenFiles    int    `json:"max_open_files"`
		MapSize         int    `json:"map_size"`
	} `json:"db"`

	BinLog struct {
		Use         bool `json:"use"`
		MaxFileSize int  `json:"max_file_size"`
		MaxFileNum  int  `json:"max_file_num"`
	} `json:"binlog"`

	//set slaveof to enable replication from master
	//empty, no replication
	SlaveOf string `json:"slaveof"`

	AccessLog string `json:"access_log"`
}

func NewConfig(data json.RawMessage) (*Config, error) {
	c := new(Config)

	err := json.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func NewConfigWithFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return NewConfig(data)
}

func (cfg *Config) NewLedisConfig() *ledis.Config {
	c := new(ledis.Config)

	c.DataDir = cfg.DataDir

	copier.Copy(&c.DB, &cfg.DB)
	copier.Copy(&c.BinLog, &cfg.BinLog)

	return c
}
