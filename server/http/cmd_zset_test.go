package http

import (
	"fmt"
	"testing"
	"time"
)

func TestZAddCommand(t *testing.T) {
	db := getTestDB()
	_, err := zaddCommand(db, "test_zadd")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zadd") {
		t.Fatal("invalid err ", err)
	}

	v, err := zaddCommand(db, "test_zadd", "10", "m")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}

}

func TestZCardCommand(t *testing.T) {
	db := getTestDB()
	_, err := zcardCommand(db, "test_zcard", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zcard") {
		t.Fatal("invalid err ", err)
	}

	v, err := zcardCommand(db, "test_zcard")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZScore(t *testing.T) {
	db := getTestDB()
	_, err := zscoreCommand(db, "test_zscore")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zscore") {
		t.Fatal("invalid err ", err)
	}
	v, err := zscoreCommand(db, "test_zscore", "m")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v != nil {
		t.Fatal("invalid result ", v)
	}
}

func TestZRemCommand(t *testing.T) {
	db := getTestDB()
	_, err := zremCommand(db, "test_zrem")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zrem") {
		t.Fatal("invalid err ", err)
	}
	v, err := zremCommand(db, "test_zrem", "m")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZIncrbyCommand(t *testing.T) {
	db := getTestDB()
	_, err := zincrbyCommand(db, "test_zincrby")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zincrby") {
		t.Fatal("invalid err ", err)
	}
	v, err := zincrbyCommand(db, "test_zincrby", "10", "m")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(string) != "10" {
		t.Fatal("invalid result ", v)
	}
}

func TestZCountCommand(t *testing.T) {
	db := getTestDB()
	_, err := zcountCommand(db, "test_zcount")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zcount") {
		t.Fatal("invalid err ", err)
	}
	v, err := zcountCommand(db, "test_zcount", "0", "1")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZRankCommand(t *testing.T) {
	db := getTestDB()
	_, err := zrankCommand(db, "test_zrank")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zrank") {
		t.Fatal("invalid err ", err)
	}
	v, err := zrankCommand(db, "test_zcount", "m")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v != nil {
		t.Fatal("invalid result ", v)
	}
}

func TestZRevrankCommand(t *testing.T) {
	db := getTestDB()
	_, err := zrevrankCommand(db, "test_zrevrank")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zrevrank") {
		t.Fatal("invalid err ", err)
	}
	v, err := zrevrankCommand(db, "test_zrevrank", "m")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v != nil {
		t.Fatal("invalid result ", v)
	}
}

func TestZRemrangebyrankCommand(t *testing.T) {
	db := getTestDB()
	_, err := zremrangebyrankCommand(db, "test_zremrangebyrank")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zremrangebyrank") {
		t.Fatal("invalid err ", err)
	}
	v, err := zremrangebyrankCommand(db, "test_zremrangebyrank", "0", "1")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZRemrangebyscore(t *testing.T) {
	db := getTestDB()
	_, err := zremrangebyscoreCommand(db, "test_zremrangebyscore")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zremrangebyscore") {
		t.Fatal("invalid err ", err)
	}
	v, err := zremrangebyscoreCommand(db, "test_zremrangebyscore", "0", "1")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZRangeCommand(t *testing.T) {
	db := getTestDB()
	_, err := zrangeCommand(db, "test_zrange")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zrange") {
		t.Fatal("invalid err ", err)
	}
	v, err := zrangeCommand(db, "test_zrange", "0", "1")
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(v.([]string)) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZRangebyscoreCommand(t *testing.T) {
	db := getTestDB()
	_, err := zrangebyscoreCommand(db, "test_zrangebyscore")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zrangebyscore") {
		t.Fatal("invalid err ", err)
	}
	v, err := zrangebyscoreCommand(db, "test_zrangebyscore", "0", "1")
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(v.([]string)) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZRevrangebyscoreCommand(t *testing.T) {
	db := getTestDB()
	_, err := zrevrangebyscoreCommand(db, "test_zrevrangebyscore")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zrevrangebyscore") {
		t.Fatal("invalid err ", err)
	}
	v, err := zrevrangebyscoreCommand(db, "test_zrevrangebyscore", "0", "1")
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(v.([]string)) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZMclearCommand(t *testing.T) {
	db := getTestDB()
	_, err := zmclearCommand(db)
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zmclear") {
		t.Fatal("invalid err ", err)
	}
	v, err := zmclearCommand(db, "test_zmclear")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 1 {
		t.Fatal("invalid result ", v)
	}
}

func TestZExpireCommand(t *testing.T) {
	db := getTestDB()
	_, err := zexpireCommand(db, "test_zexpire")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zexpire") {
		t.Fatal("invalid err ", err)
	}
	v, err := zexpireCommand(db, "test_zexpire", "10")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZExpireAtCommand(t *testing.T) {
	db := getTestDB()
	_, err := zexpireAtCommand(db, "test_zexpireat")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zexpireat") {
		t.Fatal("invalid err ", err)
	}
	expireAt := fmt.Sprintf("%d", time.Now().Unix()+10)
	v, err := zexpireAtCommand(db, "test_zexpire", expireAt)
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}

func TestZTTLCommand(t *testing.T) {
	db := getTestDB()
	_, err := zttlCommand(db, "test_zttl", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zttl") {
		t.Fatal("invalid err ", err)
	}
	v, err := zttlCommand(db, "test_zttl")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != -1 {
		t.Fatal("invalid result ", v)
	}
}

func TestZPersistCommand(t *testing.T) {
	db := getTestDB()
	_, err := zpersistCommand(db, "test_zpersist", "a")
	if err == nil || err.Error() != fmt.Sprintf(ERR_ARGUMENT_FORMAT, "zpersist") {
		t.Fatal("invalid err ", err)
	}
	v, err := zpersistCommand(db, "test_zpersist")
	if err != nil {
		t.Fatal(err.Error())
	}

	if v.(int64) != 0 {
		t.Fatal("invalid result ", v)
	}
}
