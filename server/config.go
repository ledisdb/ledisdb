package server

import (
	"encoding/json"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/replication"
	"io/ioutil"
)

type Config struct {
	Addr string `json:"addr"`

	DataDir string `json:"data_dir"`

	//if you not set db path, use data_dir
	DB ledis.Config `json:"db"`

	//if you not set relay log path, use data_dir/realy_log
	RelayLog replication.RelayLogConfig `json:"relay_log"`
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
