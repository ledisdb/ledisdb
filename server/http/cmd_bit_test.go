package http

import (
	//	"github.com/siddontang/ledisdb/ledis"
	"testing"
)

func TestBgetCommand(t *testing.T) {
	db := getTestDB()
	db.BSetBit([]byte("test_bget"), 0, 1)
	db.BSetBit([]byte("test_bget"), 1, 1)
	db.BSetBit([]byte("test_bget"), 2, 1)

	_, err := bgetCommand(db, "test_bget", "a", "b", "c")
	if err == nil || err.Error() != "ERR wrong number of arguments for 'bget' command" {
		t.Fatal("invalid err %v", err)
	}

	r, err := bgetCommand(db, "test_bget")
	if err != nil {
		t.Fatal(err.Error())
	}
	str := r.(string)
	if str != "\x07" {
		t.Fatal("wrong result of 'bget': %v", []byte(str))
	}
}

func TestBDeleteCommand(t *testing.T) {
	db := getTestDB()

	_, err := bdeleteCommand(db, "test_bdelete", "a", "b", "c")
	if err == nil || err.Error() != "ERR wrong number of arguments for 'bdelete' command" {
		t.Fatalf("invalid err %v", err)
	}

	db.BSetBit([]byte("test_bdelete"), 0, 1)
	db.BSetBit([]byte("test_bdelete"), 1, 1)
	db.BSetBit([]byte("test_bdelete"), 2, 1)
	n, err := bdeleteCommand(db, "test_bdelete")
	if err != nil {
		t.Fatal(err.Error())
	}
	if n.(int64) != 1 {
		t.Fatalf("wrong result: %v", n)
	}

	n, err = bdeleteCommand(db, "test_bdelete_not_exit")
	if err != nil {
		t.Fatal(err.Error())
	}
	if n.(int64) != 0 {
		t.Fatalf("wrong result: %v", n)
	}
}

func TestBSetbitCommand(t *testing.T) {
	db := getTestDB()
	_, err := bsetbitCommand(db, "test_bsetbit", "a", "b", "c")
	if err == nil || err.Error() != "ERR wrong number of arguments for 'bsetbit' command" {
		t.Fatalf("invalid err %v", err)
	}
	n, err := bsetbitCommand(db, "test_bsetbit", "1", "1")
	if err != nil {
		t.Fatal(err.Error())
	}
	if n.(uint8) != 0 {
		t.Fatal("wrong result: %v", n)
	}
	n, err = db.BGetBit([]byte("test_bsetbit"), 1)
	if err != nil {
		t.Fatal(err.Error())
	}
	if n.(uint8) != 1 {
		t.Fatalf("wrong result: %v", n)
	}
}

func TestBMsetbitCommand(t *testing.T) {
	db := getTestDB()
	_, err := bmsetbitCommand(db, "test_bmsetbit", "a", "b", "c")

	if err == nil || err.Error() != "ERR wrong number of arguments for 'bmsetbit' command" {
		t.Fatalf("invalid err %v", err)
	}
	n, err := bmsetbitCommand(db, "test_bmsetbit", "1", "1", "3", "1")
	if err != nil {
		t.Fatal(err.Error())
	}
	if n.(int64) != 2 {
		t.Fatalf("wrong result: %v", n)
	}
}
