package http

import (
	"fmt"
	"testing"
	"time"
)

func TestBgetCommand(t *testing.T) {
	db := getTestDB()
	db.BSetBit([]byte("test_bget"), 0, 1)
	db.BSetBit([]byte("test_bget"), 1, 1)
	db.BSetBit([]byte("test_bget"), 2, 1)

	_, err := bgetCommand(db, "test_bget", "a", "b", "c")
	if err == nil || err.Error() != "wrong number of arguments for 'bget' command" {
		t.Fatalf("invalid err %v", err)
	}

	r, err := bgetCommand(db, "test_bget")
	if err != nil {
		t.Fatal(err.Error())
	}
	str := r.(string)
	if str != "\x07" {
		t.Fatalf("wrong result of 'bget': %v", []byte(str))
	}
}

func TestBDeleteCommand(t *testing.T) {
	db := getTestDB()

	_, err := bdeleteCommand(db, "test_bdelete", "a", "b", "c")
	if err == nil || err.Error() != "wrong number of arguments for 'bdelete' command" {
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
	if err == nil || err.Error() != "wrong number of arguments for 'bsetbit' command" {
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

	if err == nil || err.Error() != "wrong number of arguments for 'bmsetbit' command" {
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

func TestBCountCommand(t *testing.T) {
	db := getTestDB()
	_, err := bcountCommand(db, "test_bcount", "a", "b", "c")
	if err == nil || err.Error() != "wrong number of arguments for 'bcount' command" {
		t.Fatalf("invalid err %v", err)
	}

	db.BSetBit([]byte("test_bcount"), 1, 1)
	db.BSetBit([]byte("test_bcount"), 3, 1)

	cnt, err := bcountCommand(db, "test_bcount", "0", "3")
	if err != nil {
		t.Fatal(err.Error())
	}
	if cnt.(int32) != 2 {
		t.Fatal("invalid value", cnt)
	}

	cnt, err = bcountCommand(db, "test_bcount", "2")

	if err != nil {
		t.Fatal(err.Error())
	}
	if cnt.(int32) != 1 {
		t.Fatal("invalid value", cnt)
	}

	cnt, err = bcountCommand(db, "test_bcount")

	if err != nil {
		t.Fatal(err.Error())
	}
	if cnt.(int32) != 2 {
		t.Fatal("invalid value", cnt)
	}
}

func TestBOptCommand(t *testing.T) {
	db := getTestDB()
	_, err := boptCommand(db, "test_bopt")
	if err == nil || err.Error() != "wrong number of arguments for 'bopt' command" {
		t.Fatalf("invalid err %v", err)
	}

	db.BSetBit([]byte("test_bopt_and_1"), 1, 1)
	db.BSetBit([]byte("test_bopt_and_2"), 1, 1)

	_, err = boptCommand(db, "and", "test_bopt_and_3", "test_bopt_and_1", "test_bopt_and_2")
	if err != nil {
		t.Fatal(err.Error())
	}

	r, _ := db.BGet([]byte("test_bopt_and_3"))
	if len(r) != 1 || r[0] != 2 {
		t.Fatalf("invalid result %v", r)
	}

	db.BSetBit([]byte("test_bopt_or_1"), 0, 1)
	db.BSetBit([]byte("test_bopt_or_1"), 1, 1)
	db.BSetBit([]byte("test_bopt_or_2"), 0, 1)
	db.BSetBit([]byte("test_bopt_or_2"), 2, 1)

	_, err = boptCommand(db, "or", "test_bopt_or_3", "test_bopt_or_1", "test_bopt_or_2")
	if err != nil {
		t.Fatal(err.Error())
	}

	r, _ = db.BGet([]byte("test_bopt_or_3"))
	if len(r) != 1 || r[0] != 7 {
		t.Fatalf("invalid result %v", r)
	}

	db.BSetBit([]byte("test_bopt_xor_1"), 0, 1)
	db.BSetBit([]byte("test_bopt_xor_1"), 1, 1)
	db.BSetBit([]byte("test_bopt_xor_2"), 0, 1)
	db.BSetBit([]byte("test_bopt_xor_2"), 2, 1)

	_, err = boptCommand(db, "xor", "test_bopt_xor_3", "test_bopt_xor_1", "test_bopt_xor_2")
	if err != nil {
		t.Fatal(err.Error())
	}

	r, _ = db.BGet([]byte("test_bopt_xor_3"))
	if len(r) != 1 || r[0] != 6 {
		t.Fatalf("invalid result %v", r)
	}

	db.BSetBit([]byte("test_bopt_not_1"), 0, 1)
	db.BSetBit([]byte("test_bopt_not_1"), 1, 0)
	_, err = boptCommand(db, "not", "test_bopt_not_2", "test_bopt_not_1")
	if err != nil {
		t.Fatal(err.Error())
	}

	r, _ = db.BGet([]byte("test_bopt_not_2"))
	if len(r) != 1 || r[0] != 2 {
		t.Fatalf("invalid result %v", r)
	}

	_, err = boptCommand(db, "invalid_opt", "abc")
	if err == nil || err.Error() != "invalid argument 'invalid_opt' for 'bopt' command" {
		t.Fatal("invalid err ", err.Error())
	}
}

func TestBExpireCommand(t *testing.T) {
	db := getTestDB()
	_, err := bexpireCommand(db, "test_bexpire", "a", "b")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "bexpire") {
		t.Fatalf("invalid err %v", err)
	}

	db.BSetBit([]byte("test_bexpire"), 1, 1)
	bexpireCommand(db, "test_bexpire", "1000")

	n, err := db.BTTL([]byte("test_bexpire"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if n == -1 {
		t.Fatal("wrong result ", n)
	}
}

func TestBExpireAtCommand(t *testing.T) {
	db := getTestDB()
	_, err := bexpireatCommand(db, "test_bexpireat", "a", "b")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "bexpireat") {
		t.Fatalf("invalid err %v", err)
	}

	db.BSetBit([]byte("test_bexpireat"), 1, 1)
	expireAt := fmt.Sprintf("%d", time.Now().Unix()+100)
	if _, err = bexpireatCommand(db, "test_bexpireat", expireAt); err != nil {
		t.Fatal(err.Error())
	}

	n, err := db.BTTL([]byte("test_bexpireat"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if n == -1 {
		t.Fatal("wrong result ", n)
	}
}

func TestBTTLCommand(t *testing.T) {
	db := getTestDB()

	_, err := bttlCommand(db, "test_bttl", "a", "b")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "bttl") {
		t.Fatalf("invalid err %v", err)
	}

	v, err := bttlCommand(db, "test_bttl")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != -1 {
		t.Fatal("invalid result ", v)
	}
}

func TestBPersistCommand(t *testing.T) {

	db := getTestDB()
	_, err := bpersistCommand(db, "test_bpersist", "a", "b")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "bpersist") {
		t.Fatalf("invalid err %v", err)
	}

	v, err := bpersistCommand(db, "test_bpersist")
	if err != nil {
		t.Fatal(err.Error())
	}
	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}
