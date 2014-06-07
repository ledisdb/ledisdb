package server

import (
	"encoding/json"
	"github.com/siddontang/ledisdb/ledis"
	"io/ioutil"
)

type Config struct {
	Addr string `json:"addr"`

	DataDir string `json:"data_dir"`

	//if you not set db path, use data_dir
	DB ledis.Config `json:"db"`

	//set slaveof to enable replication from master
	//empty, no replication
	SlaveOf string `json:"slaveof"`
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
