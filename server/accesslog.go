package server

import (
	"fmt"
	"github.com/siddontang/go-log/log"
	"strings"
)

const (
	accessTimeFormat = "2006/01/02 15:04:05"
)

type accessLog struct {
	l *log.Logger
}

func newAcessLog(baseName string) (*accessLog, error) {
	l := new(accessLog)

	h, err := log.NewTimeRotatingFileHandler(baseName, log.WhenDay, 1)
	if err != nil {
		return nil, err
	}

	l.l = log.New(h, log.Ltime)

	return l, nil
}

func (l *accessLog) Close() {
	l.l.Close()
}

func (l *accessLog) Log(remoteAddr string, usedTime int64, cmd string, args [][]byte, err error) {
	argsFormat := make([]string, len(args))
	argsI := make([]interface{}, len(args))
	for i := range args {
		argsFormat[i] = " %.24q"
		argsI[i] = args[i]
	}

	argsStr := fmt.Sprintf(strings.Join(argsFormat, ""), argsI...)

	format := `%s [%s%s] %d [%s]`

	if err == nil {
		l.l.Info(format, remoteAddr, cmd, argsStr, usedTime, "OK")
	} else {
		l.l.Info(format, remoteAddr, cmd, argsStr, usedTime, err.Error())
	}
}
