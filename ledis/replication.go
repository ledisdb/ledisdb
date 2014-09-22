package ledis

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/rpl"
	"io"
	"time"
)

const (
	maxReplLogSize = 1 * 1024 * 1024
)

var (
	ErrLogMissed = errors.New("log is pured in server")
)

func (l *Ledis) handleReplication() {
	l.rwg.Add(1)
	var rl *rpl.Log
	for {
		if err := l.r.NextCommitLog(rl); err != nil {
			if err != rpl.ErrNoBehindLog {
				log.Error("get next commit log err, %s", err.Error)
			} else {
				l.rwg.Done()
				return
			}
		} else {
			l.rbatch.Rollback()
			decodeEventBatch(l.rbatch, rl.Data)

			if err := l.rbatch.Commit(); err != nil {
				log.Error("commit log error %s", err.Error())
			} else if err = l.r.UpdateCommitID(rl.ID); err != nil {
				log.Error("update commit id error %s", err.Error())
			}
		}

	}
}

func (l *Ledis) onReplication() {
	if l.r == nil {
		return
	}

	for {
		select {
		case <-l.rc:
			l.handleReplication()
		case <-time.After(5 * time.Second):
			l.handleReplication()
		}
	}
}

func (l *Ledis) WaitReplication() error {
	l.rwg.Wait()

	return nil
}

func (l *Ledis) StoreLogsFromReader(rb io.Reader) (uint64, error) {
	if l.r == nil {
		return 0, fmt.Errorf("replication not enable")
	}

	var log *rpl.Log
	var n uint64

	for {
		if err := log.Decode(rb); err != nil {
			if err == io.EOF {
				break
			} else {
				return 0, err
			}
		}

		if err := l.r.StoreLog(log); err != nil {
			return 0, err
		}

		n = log.ID
	}

	select {
	case l.rc <- struct{}{}:
	default:
		break
	}

	return n, nil
}

func (l *Ledis) StoreLogsFromData(data []byte) (uint64, error) {
	rb := bytes.NewReader(data)

	return l.StoreLogsFromReader(rb)
}

func (l *Ledis) ReadLogsTo(startLogID uint64, w io.Writer) (n int, nextLogID uint64, err error) {
	if l.r == nil {
		// no replication log
		nextLogID = 0
		return
	}

	var firtID, lastID uint64

	firtID, err = l.r.FirstLogID()
	if err != nil {
		return
	}

	if startLogID < firtID {
		err = ErrLogMissed
		return
	}

	lastID, err = l.r.LastLogID()
	if err != nil {
		return
	}

	var log *rpl.Log
	for i := startLogID; i <= lastID; i++ {
		if err = l.r.GetLog(i, log); err != nil {
			return
		}

		if err = log.Encode(w); err != nil {
			return
		}

		nextLogID = i + 1

		n += log.Size()

		if n > maxReplLogSize {
			break
		}
	}

	return
}

// try to read events, if no events read, try to wait the new event singal until timeout seconds
func (l *Ledis) ReadLogsToTimeout(startLogID uint64, w io.Writer, timeout int) (n int, nextLogID uint64, err error) {
	n, nextLogID, err = l.ReadLogsTo(startLogID, w)
	if err != nil {
		return
	} else if n == 0 || nextLogID == 0 {
		return
	}
	//no events read
	select {
	//case <-l.binlog.Wait():
	case <-time.After(time.Duration(timeout) * time.Second):
	}
	return l.ReadLogsTo(startLogID, w)

}
