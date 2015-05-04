package server

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/siddontang/goredis"
	"github.com/siddontang/ledisdb/config"
)

func checkDataEqual(master *App, slave *App) error {
	mdb, _ := master.ldb.Select(0)
	sdb, _ := slave.ldb.Select(0)

	mkeys, _ := mdb.Scan(KV, nil, 100, true, "")
	skeys, _ := sdb.Scan(KV, nil, 100, true, "")

	if len(mkeys) != len(skeys) {
		return fmt.Errorf("keys number not equal %d != %d", len(mkeys), len(skeys))
	} else if !reflect.DeepEqual(mkeys, skeys) {
		return fmt.Errorf("keys not equal")
	} else {
		for _, k := range mkeys {
			v1, _ := mdb.Get(k)
			v2, _ := sdb.Get(k)
			if !reflect.DeepEqual(v1, v2) {
				return fmt.Errorf("value not equal")
			}
		}
	}

	return nil
}

func TestReplication(t *testing.T) {
	data_dir := "/tmp/test_replication"
	os.RemoveAll(data_dir)

	masterCfg := config.NewConfigDefault()
	masterCfg.DataDir = fmt.Sprintf("%s/master", data_dir)
	masterCfg.Addr = "127.0.0.1:11182"
	masterCfg.UseReplication = true
	masterCfg.Replication.Sync = true
	masterCfg.Replication.WaitSyncTime = 5000

	var master *App
	var slave *App
	var err error
	master, err = NewApp(masterCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()

	slaveCfg := config.NewConfigDefault()
	slaveCfg.DataDir = fmt.Sprintf("%s/slave", data_dir)
	slaveCfg.Addr = "127.0.0.1:11183"
	slaveCfg.SlaveOf = masterCfg.Addr
	slaveCfg.UseReplication = true

	slave, err = NewApp(slaveCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer slave.Close()

	go master.Run()

	time.Sleep(1 * time.Second)
	go slave.Run()

	db, _ := master.ldb.Select(0)

	value := make([]byte, 10)

	db.Set([]byte("a"), value)
	db.Set([]byte("b"), value)
	db.Set([]byte("c"), value)
	db.Set([]byte("d"), value)

	time.Sleep(1 * time.Second)
	if err = checkDataEqual(master, slave); err != nil {
		t.Fatal(err)
	}

	db.Set([]byte("a1"), value)
	db.Set([]byte("b1"), value)
	db.Set([]byte("c1"), value)
	db.Set([]byte("d1"), value)

	//time.Sleep(1 * time.Second)
	slave.ldb.WaitReplication()

	if err = checkDataEqual(master, slave); err != nil {
		t.Fatal(err)
	}

	slave.slaveof("", false, false)

	db.Set([]byte("a2"), value)
	db.Set([]byte("b2"), value)
	db.Set([]byte("c2"), value)
	db.Set([]byte("d2"), value)

	db.Set([]byte("a3"), value)
	db.Set([]byte("b3"), value)
	db.Set([]byte("c3"), value)
	db.Set([]byte("d3"), value)

	if err = checkDataEqual(master, slave); err == nil {
		t.Fatal("must error")
	}

	slave.slaveof(masterCfg.Addr, false, false)

	time.Sleep(1 * time.Second)

	if err = checkDataEqual(master, slave); err != nil {
		t.Fatal(err)
	}

	mStat, _ := master.ldb.ReplicationStat()
	sStat, _ := slave.ldb.ReplicationStat()

	if err = checkTestRole(masterCfg.Addr, []interface{}{
		[]byte("master"),
		int64(mStat.LastID),
		[]interface{}{
			[]interface{}{
				[]byte("127.0.0.1"),
				[]byte("11183"),
				[]byte(fmt.Sprintf("%d", sStat.LastID)),
			}},
	}); err != nil {
		t.Fatal(err)
	}

	if err = checkTestRole(slaveCfg.Addr, []interface{}{
		[]byte("slave"),
		[]byte("127.0.0.1"),
		int64(11182),
		[]byte("connected"),
		int64(sStat.LastID),
	}); err != nil {
		t.Fatal(err)
	}

	slave.tryReSlaveof()

	time.Sleep(1 * time.Second)

	slave.ldb.WaitReplication()

	if err = checkDataEqual(master, slave); err != nil {
		t.Fatal(err)
	}
}

func checkTestRole(addr string, checkRoles []interface{}) error {
	conn, _ := goredis.Connect(addr)
	defer conn.Close()
	roles, err := goredis.MultiBulk(conn.Do("ROLE"))
	if err != nil {
		return err
	} else if !reflect.DeepEqual(roles, checkRoles) {
		return fmt.Errorf("%v != %v", roles, checkRoles)
	}
	return nil
}
