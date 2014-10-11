package config

import (
	"bytes"
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/siddontang/go/ioutil2"
	"io"
	"io/ioutil"
)

var (
	ErrNoConfigFile = errors.New("Running without a config file")
)

const (
	DefaultAddr string = "127.0.0.1:6380"

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
	Sync             bool   `toml:"sync"`
	WaitSyncTime     int    `toml:"wait_sync_time"`
	WaitMaxSlaveAcks int    `toml:"wait_max_slave_acks"`
	ExpiredLogDays   int    `toml:"expired_log_days"`
	SyncLog          int    `toml:"sync_log"`
	Compression      bool   `toml:"compression"`
}

type SnapshotConfig struct {
	Path   string `toml:"path"`
	MaxNum int    `toml:"max_num"`
}

type Config struct {
	FileName string `toml:"-"`

	Addr string `toml:"addr"`

	HttpAddr string `toml:"http_addr"`

	SlaveOf string `toml:"slaveof"`

	Readonly bool `toml:readonly`

	DataDir string `toml:"data_dir"`

	DBName string `toml:"db_name"`

	DBPath string `toml:"db_path"`

	LevelDB LevelDBConfig `toml:"leveldb"`

	LMDB LMDBConfig `toml:"lmdb"`

	AccessLog string `toml:"access_log"`

	UseReplication bool              `toml:"use_replication"`
	Replication    ReplicationConfig `toml:"replication"`

	Snapshot SnapshotConfig `toml:"snapshot"`
}

func NewConfigWithFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	if cfg, err := NewConfigWithData(data); err != nil {
		return nil, err
	} else {
		cfg.FileName = fileName
		return cfg, nil
	}
}

func NewConfigWithData(data []byte) (*Config, error) {
	cfg := NewConfigDefault()

	_, err := toml.Decode(string(data), cfg)
	if err != nil {
		return nil, err
	}

	cfg.adjust()

	return cfg, nil
}

func NewConfigDefault() *Config {
	cfg := new(Config)

	cfg.Addr = DefaultAddr
	cfg.HttpAddr = ""

	cfg.DataDir = DefaultDataDir

	cfg.DBName = DefaultDBName

	cfg.SlaveOf = ""
	cfg.Readonly = false

	// disable access log
	cfg.AccessLog = ""

	cfg.LMDB.MapSize = 20 * 1024 * 1024
	cfg.LMDB.NoSync = true

	cfg.UseReplication = false
	cfg.Replication.WaitSyncTime = 500
	cfg.Replication.Compression = true
	cfg.Replication.WaitMaxSlaveAcks = 2
	cfg.Replication.SyncLog = 0

	cfg.adjust()

	return cfg
}

func (cfg *Config) adjust() {
	if cfg.LevelDB.CacheSize <= 0 {
		cfg.LevelDB.CacheSize = 4 * 1024 * 1024
	}

	if cfg.LevelDB.BlockSize <= 0 {
		cfg.LevelDB.BlockSize = 4 * 1024
	}

	if cfg.LevelDB.WriteBufferSize <= 0 {
		cfg.LevelDB.WriteBufferSize = 4 * 1024 * 1024
	}

	if cfg.LevelDB.MaxOpenFiles < 1024 {
		cfg.LevelDB.MaxOpenFiles = 1024
	}

	if cfg.Replication.ExpiredLogDays <= 0 {
		cfg.Replication.ExpiredLogDays = 7
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

func (cfg *Config) Rewrite() error {
	if len(cfg.FileName) == 0 {
		return ErrNoConfigFile
	}

	return cfg.DumpFile(cfg.FileName)
}
