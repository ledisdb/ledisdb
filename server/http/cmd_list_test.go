package http

import (
	"fmt"
	"testing"
	"time"
)

func TestLpushCommand(t *testing.T) {
	db := getTestDB()
	_, err := lpushCommand(db, "test_lpush")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lpush") {
		t.Fatal("invalid err ", err)
	}

	v, err := lpushCommand(db, "test_lpush", "1", "2")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 2 {
		t.Fatal("invalid result", v)
	}
}

func TestRpushCommand(t *testing.T) {
	db := getTestDB()
	_, err := rpushCommand(db, "test_rpush")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "rpush") {
		t.Fatal("invalid err ", err)
	}

	v, err := rpushCommand(db, "test_rpush", "1", "2")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 2 {
		t.Fatal("invalid result", v)
	}
}

func TestLpopCommand(t *testing.T) {
	db := getTestDB()
	_, err := lpopCommand(db, "test_lpop", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lpop") {
		t.Fatal("invalid err ", err)
	}

	v, err := lpopCommand(db, "test_lpop")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v != nil {
		t.Fatal("invalid result", v)
	}
}

func TestRpopCommand(t *testing.T) {
	db := getTestDB()
	_, err := rpopCommand(db, "test_rpop", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "rpop") {
		t.Fatal("invalid err ", err)
	}

	v, err := rpopCommand(db, "test_lpop")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v != nil {
		t.Fatal("invalid result", v)
	}
}

func TestLlenCommand(t *testing.T) {
	db := getTestDB()
	_, err := llenCommand(db, "test_llen", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "llen") {
		t.Fatal("invalid err ", err)
	}

	v, err := llenCommand(db, "test_llen")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result", v)
	}
}

func TestLindexCommand(t *testing.T) {
	db := getTestDB()
	_, err := lindexCommand(db, "test_lindex")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lindex") {
		t.Fatal("invalid err ", err)
	}
	v, err := lindexCommand(db, "test_lindex", "1")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v != nil {
		t.Fatal("invalid result", v)
	}
}

func TestLrangeCommand(t *testing.T) {
	db := getTestDB()
	_, err := lrangeCommand(db, "test_lrange")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lrange") {
		t.Fatal("invalid err ", err)
	}
	v, err := lrangeCommand(db, "test_lrange", "1", "2")
	if err != nil {
		t.Fatal(err.Error())
	}

	arr := v.([]interface{})
	if len(arr) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestLclearCommand(t *testing.T) {
	db := getTestDB()
	_, err := lclearCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lclear") {
		t.Fatal("invalid err ", err)
	}
	v, err := lclearCommand(db, "test_lclear")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestLmclearCommand(t *testing.T) {
	db := getTestDB()
	_, err := lmclearCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lmclear") {
		t.Fatal("invalid err ", err)
	}
	v, err := lmclearCommand(db, "test_lmclear")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}
}

func TestLexpireCommand(t *testing.T) {
	db := getTestDB()
	_, err := lexpireCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lexpire") {
		t.Fatal("invalid err ", err)
	}
	v, err := lexpireCommand(db, "test_lexpire", "10")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestLexpireAtCommand(t *testing.T) {
	db := getTestDB()
	_, err := lexpireAtCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lexpireat") {
		t.Fatal("invalid err ", err)
	}
	expireAt := fmt.Sprintf("%d", time.Now().Unix())
	v, err := lexpireCommand(db, "test_lexpireat", expireAt)
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestLTTLCommand(t *testing.T) {
	db := getTestDB()
	_, err := lttlCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lttl") {
		t.Fatal("invalid err ", err)
	}
	v, err := lttlCommand(db, "test_lttl")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != -1 {
		t.Fatal("invalid result ", v)
	}
}

func TestLpersistCommand(t *testing.T) {
	db := getTestDB()
	_, err := lpersistCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "lpersist") {
		t.Fatal("invalid err ", err)
	}

	v, err := lpersistCommand(db, "test_lpersist")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}
