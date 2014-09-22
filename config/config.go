package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

type Size int

const (
	DefaultAddr     string = "127.0.0.1:6380"
	DefaultHttpAddr string = "127.0.0.1:11181"

	DefaultDBName string = "goleveldb"

	DefaultDataDir string = "./var"
)

type LevelDBConfig struct {
	Compression     bool `toml:"compression"`
	BlockSize       int  `toml:"block_size"`
	WriteBufferSize int  `toml:"write_buffer_size"`
	CacheSize       int  `toml:"cache_size"`
	MaxOpenFiles    int  `toml:"max_open_files"`
}

type LMDBConfig struct {
	MapSize int  `toml:"map_size"`
	NoSync  bool `toml:"nosync"`
}

type ReplicationConfig struct {
	Use            bool   `toml:"use"`
	Path           string `toml:"path"`
	ExpiredLogDays int    `toml:"expired_log_days"`
}

type Config struct {
	Addr string `toml:"addr"`

	HttpAddr string `toml:"http_addr"`

	SlaveOf string `toml:"slaveof"`

	DataDir string `toml:"data_dir"`

	DBName string `toml:"db_name"`

	DBPath string `toml:"db_path"`

	LevelDB LevelDBConfig `toml:"leveldb"`

	LMDB LMDBConfig `toml:"lmdb"`

	AccessLog string `toml:"access_log"`

	Replication ReplicationConfig `toml:"replication"`
}

func NewConfigWithFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return NewConfigWithData(data)
}

func NewConfigWithData(data []byte) (*Config, error) {
	cfg := NewConfigDefault()

	_, err := toml.Decode(string(data), cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func NewConfigDefault() *Config {
	cfg := new(Config)

	cfg.Addr = DefaultAddr
	cfg.HttpAddr = DefaultHttpAddr

	cfg.DataDir = DefaultDataDir

	cfg.DBName = DefaultDBName

	cfg.SlaveOf = ""

	// disable access log
	cfg.AccessLog = ""

	cfg.LMDB.MapSize = 20 * 1024 * 1024
	cfg.LMDB.NoSync = true

	return cfg
}

func (cfg *LevelDBConfig) Adjust() {
	if cfg.CacheSize <= 0 {
		cfg.CacheSize = 4 * 1024 * 1024
	}

	if cfg.BlockSize <= 0 {
		cfg.BlockSize = 4 * 1024
	}

	if cfg.WriteBufferSize <= 0 {
		cfg.WriteBufferSize = 4 * 1024 * 1024
	}

	if cfg.MaxOpenFiles < 1024 {
		cfg.MaxOpenFiles = 1024
	}
}
