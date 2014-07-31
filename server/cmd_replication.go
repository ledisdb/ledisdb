package server

import (
	"encoding/binary"
	"fmt"
	"github.com/siddontang/go-snappy/snappy"
	"github.com/siddontang/ledisdb/ledis"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func slaveofCommand(c *client) error {
	args := c.args

	if len(args) != 2 {
		return ErrCmdParams
	}

	masterAddr := ""

	if strings.ToLower(ledis.String(args[0])) == "no" &&
		strings.ToLower(ledis.String(args[1])) == "one" {
		//stop replication, use master = ""
	} else {
		if _, err := strconv.ParseInt(ledis.String(args[1]), 10, 16); err != nil {
			return err
		}

		masterAddr = fmt.Sprintf("%s:%s", args[0], args[1])
	}

	if err := c.app.slaveof(masterAddr); err != nil {
		return err
	}

	c.resp.writeStatus(OK)

	return nil
}

func fullsyncCommand(c *client) error {
	//todo, multi fullsync may use same dump file
	dumpFile, err := ioutil.TempFile(c.app.cfg.DataDir, "dump_")
	if err != nil {
		return err
	}

	if err = c.app.ldb.Dump(dumpFile); err != nil {
		return err
	}

	st, _ := dumpFile.Stat()
	n := st.Size()

	dumpFile.Seek(0, os.SEEK_SET)

	c.resp.writeBulkFrom(n, dumpFile)

	name := dumpFile.Name()
	dumpFile.Close()

	os.Remove(name)

	return nil
}

var reserveInfoSpace = make([]byte, 16)

func syncCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	var logIndex int64
	var logPos int64
	var err error
	logIndex, err = ledis.StrInt64(args[0], nil)
	if err != nil {
		return ErrCmdParams
	}

	logPos, err = ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrCmdParams
	}

	c.syncBuf.Reset()

	//reserve space to write master info
	if _, err := c.syncBuf.Write(reserveInfoSpace); err != nil {
		return err
	}

	m := &ledis.MasterInfo{logIndex, logPos}

	if _, err := c.app.ldb.ReadEventsTo(m, &c.syncBuf); err != nil {
		return err
	} else {
		buf := c.syncBuf.Bytes()

		binary.BigEndian.PutUint64(buf[0:], uint64(m.LogFileIndex))
		binary.BigEndian.PutUint64(buf[8:], uint64(m.LogPos))

		if len(c.compressBuf) < snappy.MaxEncodedLen(len(buf)) {
			c.compressBuf = make([]byte, snappy.MaxEncodedLen(len(buf)))
		}

		if buf, err = snappy.Encode(c.compressBuf, buf); err != nil {
			return err
		}

		c.resp.writeBulk(buf)
	}

	return nil
}

func init() {
	register("slaveof", slaveofCommand)
	register("fullsync", fullsyncCommand)
	register("sync", syncCommand)
}
