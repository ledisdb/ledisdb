package http

import (
	"fmt"
	"testing"
	"time"
)

func TestGetCommand(t *testing.T) {
	db := getTestDB()
	_, err := getCommand(db, "test_get", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "get") {
		t.Fatal("invalid err ", err)
	}

	err = db.Set([]byte("test_get"), []byte("v"))
	if err != nil {
		t.Fatal(err.Error())
	}
	v, err := getCommand(db, "test_get")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(string) != "v" {
		t.Fatalf("invalid result %v", v)
	}

}

func TestSetCommand(t *testing.T) {
	db := getTestDB()
	_, err := setCommand(db, "test_set")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "set") {
		t.Fatal("invalid err ", err)
	}
	v, err := setCommand(db, "test_set", "v")
	if err != nil {
		t.Fatal(err.Error())
	}
	r := v.([]interface{})
	if len(r) != 2 {
		t.Fatalf("invalid result %v", v)
	}

	if r[0].(bool) != true {
		t.Fatalf("invalid result %v", r[0])
	}

	if r[1].(string) != "OK" {
		t.Fatalf("invalid result %v", r[1])
	}
}

func TestGetsetCommand(t *testing.T) {
	db := getTestDB()
	_, err := getsetCommand(db, "test_getset")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "getset") {
		t.Fatal("invalid err ", err)
	}

	v, err := getsetCommand(db, "test_getset", "v")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v != nil {
		t.Fatal("invalid result ", v)
	}
}

func TestSetnxCommand(t *testing.T) {
	db := getTestDB()
	_, err := setnxCommand(db, "test_setnx")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "setnx") {
		t.Fatal("invalid err ", err)
	}
	v, err := setnxCommand(db, "test_setnx", "v")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}
}

func TestExistsCommand(t *testing.T) {
	db := getTestDB()
	_, err := existsCommand(db, "test_exists", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "exists") {
		t.Fatal("invalid err ", err)
	}
	v, err := existsCommand(db, "test_exists")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestIncrCommand(t *testing.T) {
	db := getTestDB()
	_, err := incrCommand(db, "test_incr", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "incr") {
		t.Fatal("invalid err ", err)
	}
	v, err := incrCommand(db, "test_incr")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}

}

func TestDecrCommand(t *testing.T) {
	db := getTestDB()
	_, err := decrCommand(db, "test_decr", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "decr") {
		t.Fatal("invalid err ", err)
	}

	v, err := decrCommand(db, "test_decr")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != -1 {
		t.Fatal("invalid result ", v)
	}
}

func TestDelCommand(t *testing.T) {
	db := getTestDB()
	_, err := delCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "del") {
		t.Fatal("invalid err ", err)
	}

	v, err := delCommand(db, "test_del")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}
}

func TestMsetCommand(t *testing.T) {
	db := getTestDB()
	_, err := msetCommand(db, "a", "b", "c")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "mset") {
		t.Fatal("invalid err ", err)
	}

	v, err := msetCommand(db, "test_mset", "v")

	if err != nil {
		t.Fatal(err.Error())
	}
	r := v.([]interface{})
	if len(r) != 2 {
		t.Fatal("invalid result ", v)
	}
	if r[0].(bool) != true {
		t.Fatal("invalid result ", r[0])
	}

	if r[1].(string) != "OK" {
		t.Fatal("invalid result ", r[1])
	}
}

func TestMgetCommand(t *testing.T) {
	db := getTestDB()
	_, err := mgetCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "mget") {
		t.Fatal("invalid err ", err)
	}

	v, err := mgetCommand(db, "test_mget")

	if err != nil {
		t.Fatal(err.Error())
	}
	arr := v.([]interface{})
	if arr[0] != nil {
		t.Fatal("invalid result ", arr)
	}
}

func TestExpireCommand(t *testing.T) {
	db := getTestDB()
	_, err := expireCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "expire") {
		t.Fatal("invalid err ", err)
	}
	v, err := expireCommand(db, "test_expire", "10")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestExpireAtCommand(t *testing.T) {
	db := getTestDB()
	_, err := expireAtCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "expireat") {
		t.Fatal("invalid err ", err)
	}

	expireAt := fmt.Sprintf("%d", time.Now().Unix()+10)
	v, err := expireAtCommand(db, "test_expireat", expireAt)

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestTTLCommand(t *testing.T) {
	db := getTestDB()
	_, err := ttlCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "ttl") {
		t.Fatal("invalid err ", err)
	}

	v, err := ttlCommand(db, "test_ttl")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != -1 {
		t.Fatal("invalid result ", v)
	}
}

func TestPersistCommand(t *testing.T) {
	db := getTestDB()
	_, err := persistCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "persist") {
		t.Fatal("invalid err ", err)
	}

	v, err := persistCommand(db, "test_persist")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}
