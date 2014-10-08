package config

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"github.com/siddontang/go/ioutil2"
	"io"
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
	Path             string `toml:"path"`
	ExpiredLogDays   int    `toml:"expired_log_days"`
	Sync             bool   `toml:"sync"`
	WaitSyncTime     int    `toml:"wait_sync_time"`
	WaitMaxSlaveAcks int    `toml:"wait_max_slave_acks"`
	Compression      bool   `toml:"compression"`
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

	UseReplication bool              `toml:"use_replication"`
	Replication    ReplicationConfig `toml:"replication"`
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

	cfg.Replication.WaitSyncTime = 1
	cfg.Replication.Compression = true
	cfg.Replication.WaitMaxSlaveAcks = 2

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

func (cfg *Config) Dump(w io.Writer) error {
	e := toml.NewEncoder(w)
	e.Indent = ""
	return e.Encode(cfg)
}

func (cfg *Config) DumpFile(fileName string) error {
	var b bytes.Buffer

	if err := cfg.Dump(&b); err != nil {
		return err
	}

	return ioutil2.WriteFileAtomic(fileName, b.Bytes(), 0644)
}
