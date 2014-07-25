package ledis

import (
	"github.com/siddontang/copier"
	"github.com/siddontang/ledisdb/store"
	"path"
)

type Config struct {
	DataDir string `json:"data_dir"`

	DB struct {
		Compression     bool `json:"compression"`
		BlockSize       int  `json:"block_size"`
		WriteBufferSize int  `json:"write_buffer_size"`
		CacheSize       int  `json:"cache_size"`
		MaxOpenFiles    int  `json:"max_open_files"`
		MapSize         int  `json:"map_size"`
	} `json:"db"`

	BinLog struct {
		Use         bool `json:"use"`
		MaxFileSize int  `json:"max_file_size"`
		MaxFileNum  int  `json:"max_file_num"`
	} `json:"binlog"`
}

func (cfg *Config) NewDBConfig() *store.Config {
	dbPath := path.Join(cfg.DataDir, "data")

	dbCfg := new(store.Config)
	copier.Copy(dbCfg, &cfg.DB)
	dbCfg.Path = dbPath
	return dbCfg
}

func (cfg *Config) NewBinLogConfig() *BinLogConfig {
	binLogPath := path.Join(cfg.DataDir, "bin_log")
	c := new(BinLogConfig)
	copier.Copy(c, &cfg.BinLog)
	c.Path = binLogPath
	return c
}
