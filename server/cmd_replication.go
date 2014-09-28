package server

import (
	"fmt"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/ledisdb/ledis"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func slaveofCommand(c *client) error {
	args := c.args

	if len(args) != 2 || len(args) != 3 {
		return ErrCmdParams
	}

	masterAddr := ""
	restart := false

	if strings.ToLower(hack.String(args[0])) == "no" &&
		strings.ToLower(hack.String(args[1])) == "one" {
		//stop replication, use master = ""
		if len(args) != 2 {
			return ErrCmdParams
		}
	} else {
		if _, err := strconv.ParseInt(hack.String(args[1]), 10, 16); err != nil {
			return err
		}

		masterAddr = fmt.Sprintf("%s:%s", args[0], args[1])

		if len(args) == 3 && strings.ToLower(hack.String(args[2])) == "restart" {
			restart = true
		}
	}

	if err := c.app.slaveof(masterAddr, restart); err != nil {
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

func syncCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	var logId uint64
	var err error

	if logId, err = ledis.StrUint64(args[0], nil); err != nil {
		return ErrCmdParams
	}

	c.lastLogID = logId - 1

	if c.ack != nil && logId > c.ack.id {
		asyncNotifyUint64(c.ack.ch, logId)
		c.ack = nil
	}

	c.syncBuf.Reset()

	if _, _, err := c.app.ldb.ReadLogsToTimeout(logId, &c.syncBuf, 30); err != nil {
		return err
	} else {
		buf := c.syncBuf.Bytes()

		c.resp.writeBulk(buf)
	}

	c.app.addSlave(c)

	return nil
}

func init() {
	register("slaveof", slaveofCommand)
	register("fullsync", fullsyncCommand)
	register("sync", syncCommand)
}
