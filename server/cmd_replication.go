package server

import (
	"encoding/binary"
	"fmt"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/ledis"
	"net"
	"strconv"
	"strings"
	"time"
)

func slaveofCommand(c *client) error {
	args := c.args

	if len(args) != 2 && len(args) != 3 {
		return ErrCmdParams
	}

	masterAddr := ""
	restart := false
	readonly := false

	if strings.ToLower(hack.String(args[0])) == "no" &&
		strings.ToLower(hack.String(args[1])) == "one" {
		//stop replication, use master = ""
		if len(args) == 3 && strings.ToLower(hack.String(args[2])) == "readonly" {
			readonly = true
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

	if err := c.app.slaveof(masterAddr, restart, readonly); err != nil {
		return err
	}

	c.resp.writeStatus(OK)

	return nil
}

func fullsyncCommand(c *client) error {
	args := c.args
	needNew := false
	if len(args) == 1 && strings.ToLower(hack.String(args[0])) == "new" {
		needNew = true
	}

	var s *snapshot
	var err error
	var t time.Time

	dumper := c.app.ldb

	if needNew {
		s, t, err = c.app.snap.Create(dumper)
	} else {
		if s, t, err = c.app.snap.OpenLatest(); err != nil {
			return err
		} else if s == nil {
			s, t, err = c.app.snap.Create(dumper)
		} else {
			gap := time.Duration(c.app.cfg.Replication.ExpiredLogDays*24*3600) * time.Second / 2
			minT := time.Now().Add(-gap)

			//snapshot is too old
			if t.Before(minT) {
				s.Close()
				s, t, err = c.app.snap.Create(dumper)
			}
		}
	}

	if err != nil {
		return err
	}

	n := s.Size()

	c.resp.writeBulkFrom(n, s)

	s.Close()

	return nil
}

var dummyBuf = make([]byte, 8)

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

	lastLogID := logId - 1

	stat, err := c.app.ldb.ReplicationStat()
	if err != nil {
		return err
	}

	if lastLogID > stat.LastID {
		return fmt.Errorf("invalid sync logid %d > %d + 1", logId, stat.LastID)
	}

	c.lastLogID.Set(lastLogID)

	if lastLogID == stat.LastID {
		c.app.slaveAck(c)
	}

	c.syncBuf.Reset()

	c.syncBuf.Write(dummyBuf)

	if _, _, err := c.app.ldb.ReadLogsToTimeout(logId, &c.syncBuf, 30, c.app.quit); err != nil {
		return err
	} else {
		buf := c.syncBuf.Bytes()

		stat, err = c.app.ldb.ReplicationStat()
		if err != nil {
			return err
		}

		binary.BigEndian.PutUint64(buf, stat.LastID)

		c.resp.writeBulk(buf)
	}

	return nil
}

//inner command, only for replication
//REPLCONF <option> <value> <option> <value> ...
func replconfCommand(c *client) error {
	args := c.args
	if len(args)%2 != 0 {
		return ErrCmdParams
	}

	//now only support "listening-port"
	for i := 0; i < len(args); i += 2 {
		switch strings.ToLower(hack.String(args[i])) {
		case "listening-port":
			var host string
			var err error
			if _, err = num.ParseUint16(hack.String(args[i+1])); err != nil {
				return err
			}
			if host, _, err = net.SplitHostPort(c.remoteAddr); err != nil {
				return err
			} else {
				c.slaveListeningAddr = net.JoinHostPort(host, hack.String(args[i+1]))
			}

			c.app.addSlave(c)
		default:
			return ErrSyntax
		}
	}

	c.resp.writeStatus(OK)
	return nil
}

func init() {
	register("slaveof", slaveofCommand)
	register("fullsync", fullsyncCommand)
	register("sync", syncCommand)
	register("replconf", replconfCommand)
}
