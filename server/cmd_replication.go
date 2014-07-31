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

func slaveofCommand(req *requestContext) error {
	args := req.args

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

	if err := req.app.slaveof(masterAddr); err != nil {
		return err
	}

	req.resp.writeStatus(OK)

	return nil
}

func fullsyncCommand(req *requestContext) error {
	//todo, multi fullsync may use same dump file
	dumpFile, err := ioutil.TempFile(req.app.cfg.DataDir, "dump_")
	if err != nil {
		return err
	}

	if err = req.app.ldb.Dump(dumpFile); err != nil {
		return err
	}

	st, _ := dumpFile.Stat()
	n := st.Size()

	dumpFile.Seek(0, os.SEEK_SET)

	req.resp.writeBulkFrom(n, dumpFile)

	name := dumpFile.Name()
	dumpFile.Close()

	os.Remove(name)

	return nil
}

var reserveInfoSpace = make([]byte, 16)

func syncCommand(req *requestContext) error {
	args := req.args
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

	req.syncBuf.Reset()

	//reserve space to write master info
	if _, err := req.syncBuf.Write(reserveInfoSpace); err != nil {
		return err
	}

	m := &ledis.MasterInfo{logIndex, logPos}

	if _, err := req.app.ldb.ReadEventsTo(m, &req.syncBuf); err != nil {
		return err
	} else {
		buf := req.syncBuf.Bytes()

		binary.BigEndian.PutUint64(buf[0:], uint64(m.LogFileIndex))
		binary.BigEndian.PutUint64(buf[8:], uint64(m.LogPos))

		if len(req.compressBuf) < snappy.MaxEncodedLen(len(buf)) {
			req.compressBuf = make([]byte, snappy.MaxEncodedLen(len(buf)))
		}

		if buf, err = snappy.Encode(req.compressBuf, buf); err != nil {
			return err
		}

		req.resp.writeBulk(buf)
	}

	return nil
}

func init() {
	register("slaveof", slaveofCommand)
	register("fullsync", fullsyncCommand)
	register("sync", syncCommand)
}
