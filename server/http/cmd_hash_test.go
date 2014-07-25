package http

import (
	"fmt"
	"testing"
	"time"
)

func TestHSetCommand(t *testing.T) {
	db := getTestDB()
	_, err := hsetCommand(db, "test_hset")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hset") {
		t.Fatalf("invalid err %v", err)
	}

	n, err := hsetCommand(db, "test_hset", "f", "v")
	if err != nil {
		t.Fatal(err)
	}
	if n.(int64) != 1 {
		t.Fatal("invalid result ", n)
	}
	v, err := db.HGet([]byte("test_hset"), []byte("f"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if string(v) != "v" {
		t.Fatalf("invalid result %s", v)
	}

}

func TestHGetCommand(t *testing.T) {
	db := getTestDB()
	_, err := hgetCommand(db, "test_hget")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hget") {
		t.Fatalf("invalid err %v", err)
	}
	if _, err := db.HSet([]byte("test_hget"), []byte("f"), []byte("v")); err != nil {
		t.Fatal(err.Error())
	}

	v, err := hgetCommand(db, "test_hget", "f")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(string) != "v" {
		t.Fatal("invalid result ", v)
	}

}

func TestHExistsCommand(t *testing.T) {
	db := getTestDB()
	_, err := hexistsCommand(db, "test_hexists")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hexists") {
		t.Fatalf("invalid err %v", err)
	}

	_, err = db.HSet([]byte("test_hexists"), []byte("f"), []byte("v"))
	if err != nil {
		t.Fatal(err.Error())
	}

	v, err := hexistsCommand(db, "test_hexists", "f")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}

}

func TestHDelCommand(t *testing.T) {
	db := getTestDB()
	_, err := hdelCommand(db, "test_hdel")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hdel") {
		t.Fatalf("invalid err %v", err)
	}

	_, err = db.HSet([]byte("test_hdel"), []byte("f"), []byte("v"))
	if err != nil {
		t.Fatal(err.Error())
	}

	v, err := hdelCommand(db, "test_hdel", "f")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}

	r, err := db.HGet([]byte("test_hdel"), []byte("f"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if r != nil {
		t.Fatalf("invalid result %v", r)
	}
}

func TestHLenCommand(t *testing.T) {
	db := getTestDB()
	_, err := hlenCommand(db, "test_hlen", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hlen") {
		t.Fatalf("invalid err %v", err)
	}

	_, err = db.HSet([]byte("test_hlen"), []byte("f1"), []byte("v1"))
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = db.HSet([]byte("test_hlen"), []byte("f2"), []byte("v2"))

	if err != nil {
		t.Fatal(err.Error())
	}

	v, err := hlenCommand(db, "test_hlen")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 2 {
		t.Fatal("invalid result ", v)
	}
}

func TestHIncrbyCommand(t *testing.T) {
	db := getTestDB()
	_, err := hincrbyCommand(db, "test_hincrby")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hincrby") {
		t.Fatalf("invalid err %v", err)
	}
	_, err = db.HSet([]byte("test_hincrby"), []byte("f"), []byte("10"))
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = hincrbyCommand(db, "test_hincrby", "f", "x")
	if err != ErrValue {
		t.Fatal("invalid err ", err)
	}

	v, err := hincrbyCommand(db, "test_hincrby", "f", "10")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 20 {
		t.Fatal("invalid result ", v)
	}
}

func TestHMsetCommand(t *testing.T) {
	db := getTestDB()
	_, err := hmsetCommand(db, "test_hmset")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hmset") {
		t.Fatalf("invalid err %v", err)
	}

	_, err = hmsetCommand(db, "test_hmset", "f1", "v1", "f2", "v2")
	if err != nil {
		t.Fatal(err.Error())
	}

	v, err := db.HGet([]byte("test_hmset"), []byte("f1"))

	if err != nil {
		t.Fatal(err.Error())
	}
	if string(v) != "v1" {
		t.Fatalf("invalid result %s", v)
	}

	v, err = db.HGet([]byte("test_hmset"), []byte("f2"))

	if err != nil {
		t.Fatal(err.Error())
	}
	if string(v) != "v2" {
		t.Fatalf("invalid result %s", v)
	}
}

func TestHMgetCommand(t *testing.T) {
	db := getTestDB()
	_, err := hmgetCommand(db, "test_hmget")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hmget") {
		t.Fatalf("invalid err %v", err)
	}

	_, err = db.HSet([]byte("test_hmget"), []byte("f1"), []byte("v1"))
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = db.HSet([]byte("test_hmget"), []byte("f2"), []byte("v2"))
	if err != nil {
		t.Fatal(err.Error())
	}

	v, err := hmgetCommand(db, "test_hmget", "f1", "f2")

	if err != nil {
		t.Fatal(err.Error())
	}
	arr := v.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("invalid arr %v", arr)
	}
	if arr[0].(string) != "v1" {
		t.Fatal("invalid result ", v)
	}

	if arr[1].(string) != "v2" {
		t.Fatal("invalid result ", v)
	}
}

func TestHGetallCommand(t *testing.T) {
	db := getTestDB()
	_, err := hgetallCommand(db, "test_hgetall", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hgetall") {
		t.Fatalf("invalid err %v", err)
	}

	_, err = db.HSet([]byte("test_hgetall"), []byte("f"), []byte("v"))
	if err != nil {
		t.Fatal(err.Error())
	}
	v, err := hgetallCommand(db, "test_hgetall")
	if err != nil {
		t.Fatal(err.Error())
	}
	m := v.(map[string]string)
	if m["f"] != "v" {
		t.Fatal("invalid result ", v)
	}
}

func TestHKeysCommand(t *testing.T) {
	db := getTestDB()
	_, err := hkeysCommand(db, "test_hkeys", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hkeys") {
		t.Fatalf("invalid err %v", err)
	}
	_, err = db.HSet([]byte("test_hkeys"), []byte("f"), []byte("v"))
	if err != nil {
		t.Fatal(err.Error())
	}
	v, err := hkeysCommand(db, "test_hkeys")
	if err != nil {
		t.Fatal(err.Error())
	}
	arr := v.([]string)
	if arr[0] != "f" {
		t.Fatal("invalid result ", v)
	}

}

func TestHClearCommand(t *testing.T) {
	db := getTestDB()
	_, err := hclearCommand(db, "test_hclear", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hclear") {
		t.Fatalf("invalid err %v", err)
	}

	_, err = db.HSet([]byte("test_hclear"), []byte("f"), []byte("v"))
	if err != nil {
		t.Fatal(err.Error())
	}

	v, err := hclearCommand(db, "test_hclear")

	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}
}

func TestHMclearCommand(t *testing.T) {
	db := getTestDB()
	_, err := hmclearCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hmclear") {
		t.Fatalf("invalid err %v", err)
	}

	v, err := hmclearCommand(db, "test_hmclear1", "test_hmclear2")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 2 {
		t.Fatal("invalid result ", v)
	}
}

func TestHExpireCommand(t *testing.T) {
	db := getTestDB()
	_, err := hexpireCommand(db, "test_hexpire")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hexpire") {
		t.Fatalf("invalid err %v", err)
	}
	v, err := hexpireCommand(db, "test_hexpire", "10")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestHExpireAtCommand(t *testing.T) {
	db := getTestDB()
	_, err := hexpireAtCommand(db, "test_hexpireat")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hexpireat") {
		t.Fatalf("invalid err %v", err)
	}

	expireAt := fmt.Sprintf("%d", time.Now().Unix()+10)
	v, err := hexpireCommand(db, "test_hexpireat", expireAt)

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestHTTLCommand(t *testing.T) {
	db := getTestDB()
	_, err := httlCommand(db, "test_httl", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "httl") {
		t.Fatalf("invalid err %v", err)
	}

	v, err := httlCommand(db, "test_httl")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != -1 {
		t.Fatal("invalid result ", v)
	}
}

func TestHPersistCommand(t *testing.T) {
	db := getTestDB()
	_, err := hpersistCommand(db, "test_hpersist", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "hpersist") {
		t.Fatalf("invalid err %v", err)
	}

	v, err := hpersistCommand(db, "test_hpersist")

	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}
