package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/store"
)

var configPath = flag.String("config", "", "ledisdb config file")
var dataDir = flag.String("data_dir", "", "ledisdb base data dir")
var dbName = flag.String("db_name", "", "select a db to use, it will overwrite the config's db name")

func main() {
	flag.Parse()

	if len(*configPath) == 0 {
		println("need ledis config file")
		return
	}

	cfg, err := config.NewConfigWithFile(*configPath)
	if err != nil {
		println(err.Error())
		return
	}

	if len(*dataDir) > 0 {
		cfg.DataDir = *dataDir
	}

	if len(*dbName) > 0 {
		cfg.DBName = *dbName
	}

	db, err := store.Open(cfg)
	if err != nil {
		println(err.Error())
		return
	}

	// upgrade: ttl time key 101 to ttl time key 103

	wb := db.NewWriteBatch()

	for i := uint8(0); i < ledis.MaxDBNumber; i++ {
		minK, maxK := oldKeyPair(i)

		it := db.RangeIterator(minK, maxK, store.RangeROpen)
		num := 0
		for ; it.Valid(); it.Next() {
			dt, k, t, err := decodeOldKey(i, it.RawKey())
			if err != nil {
				continue
			}

			newKey := encodeNewKey(i, dt, k, t)

			wb.Put(newKey, it.RawValue())
			wb.Delete(it.RawKey())
			num++
			if num%1024 == 0 {
				if err := wb.Commit(); err != nil {
					fmt.Printf("commit error :%s\n", err.Error())
				}
			}
		}
		it.Close()

		if err := wb.Commit(); err != nil {
			fmt.Printf("commit error :%s\n", err.Error())
		}
	}
}

func oldKeyPair(index uint8) ([]byte, []byte) {
	minB := make([]byte, 11)
	minB[0] = index
	minB[1] = ledis.ObsoleteExpTimeType
	minB[2] = 0

	maxB := make([]byte, 11)
	maxB[0] = index
	maxB[1] = ledis.ObsoleteExpTimeType
	maxB[2] = 255

	return minB, maxB
}

func decodeOldKey(index uint8, tk []byte) (byte, []byte, int64, error) {
	if len(tk) < 11 || tk[0] != index || tk[1] != ledis.ObsoleteExpTimeType {
		return 0, nil, 0, fmt.Errorf("invalid exp time key")
	}

	return tk[2], tk[11:], int64(binary.BigEndian.Uint64(tk[3:])), nil
}

func encodeNewKey(index uint8, dataType byte, key []byte, when int64) []byte {
	buf := make([]byte, len(key)+11)

	buf[0] = index
	buf[1] = ledis.ExpTimeType
	pos := 2

	binary.BigEndian.PutUint64(buf[pos:], uint64(when))
	pos += 8

	buf[pos] = dataType
	pos++

	copy(buf[pos:], key)

	return buf
}
