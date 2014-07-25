package http

import (
	"fmt"
	"testing"
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
