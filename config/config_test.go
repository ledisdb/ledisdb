package config

import (
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	dstCfg := new(Config)
	dstCfg.Addr = "127.0.0.1:6380"
	dstCfg.HttpAddr = "127.0.0.1:11181"
	dstCfg.DataDir = "/tmp/ledis_server"
	dstCfg.DBName = "leveldb"

	dstCfg.LevelDB.Compression = false
	dstCfg.LevelDB.BlockSize = 32768
	dstCfg.LevelDB.WriteBufferSize = 67108864
	dstCfg.LevelDB.CacheSize = 524288000
	dstCfg.LevelDB.MaxOpenFiles = 1024
	dstCfg.LMDB.MapSize = 524288000
	dstCfg.LMDB.NoSync = true

	cfg, err := NewConfigWithFile("./config.toml")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(dstCfg, cfg) {
		t.Fatal("parse toml error")
	}
}
