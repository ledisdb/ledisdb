package server

import (
	"encoding/json"
	"github.com/siddontang/ledisdb/ledis"
	"io/ioutil"
)

type Config struct {
	Addr string `json:"addr"`

	DB ledis.Config `json:"db"`
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
