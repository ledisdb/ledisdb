package server

import (
	"bytes"
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

type info struct {
	sync.Mutex

	app *App

	Server struct {
		OS         string
		ProceessId int
	}

	Clients struct {
		ConnectedClients int64
	}

	Persistence struct {
		DBName string
	}
}

func newInfo(app *App) (i *info, err error) {
	i = new(info)

	i.app = app

	i.Server.OS = runtime.GOOS
	i.Server.ProceessId = os.Getpid()

	if app.cfg.DBName != "" {
		i.Persistence.DBName = app.cfg.DBName
	} else {
		i.Persistence.DBName = config.DefaultDBName
	}

	return i, nil
}

func (i *info) addClients(delta int64) {
	atomic.AddInt64(&i.Clients.ConnectedClients, delta)
}
func (i *info) Close() {

}

func getMemoryHuman(m uint64) string {
	if m > GB {
		return fmt.Sprintf("%dG", m/GB)
	} else if m > MB {
		return fmt.Sprintf("%dM", m/MB)
	} else if m > KB {
		return fmt.Sprintf("%dK", m/KB)
	} else {
		return fmt.Sprintf("%d", m)
	}
}

func (i *info) Dump(section string) []byte {
	buf := &bytes.Buffer{}
	switch strings.ToLower(section) {
	case "":
		i.dumpAll(buf)
	case "server":
		i.dumpServer(buf)
	case "client":
		i.dumpClients(buf)
	case "cpu":
		i.dumpCPU(buf)
	case "mem":
		i.dumpMem(buf)
	case "persistence":
		i.dumpPersistence(buf)
	case "goroutine":
		i.dumpGoroutine(buf)
	default:
		buf.WriteString(fmt.Sprintf("# %s\r\n", section))
	}

	return buf.Bytes()
}

type infoPair struct {
	Key   string
	Value interface{}
}

func (i *info) dumpAll(buf *bytes.Buffer) {
	i.dumpServer(buf)
	buf.Write(Delims)
	i.dumpPersistence(buf)
	buf.Write(Delims)
	i.dumpClients(buf)
	buf.Write(Delims)
	i.dumpCPU(buf)
	buf.Write(Delims)
	i.dumpMem(buf)
	buf.Write(Delims)
	i.dumpGoroutine(buf)
}

func (i *info) dumpServer(buf *bytes.Buffer) {
	buf.WriteString("# Server\r\n")

	i.dumpPairs(buf, infoPair{"os", i.Server.OS},
		infoPair{"process_id", i.Server.ProceessId},
		infoPair{"addr", i.app.cfg.Addr},
		infoPair{"http_addr", i.app.cfg.HttpAddr})
}

func (i *info) dumpClients(buf *bytes.Buffer) {
	buf.WriteString("# Client\r\n")

	i.dumpPairs(buf, infoPair{"client_num", i.Clients.ConnectedClients})
}

func (i *info) dumpCPU(buf *bytes.Buffer) {
	buf.WriteString("# CPU\r\n")

	var rusage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage); err != nil {
		return
	}

	i.dumpPairs(buf, infoPair{"cpu_sys", rusage.Stime.Usec},
		infoPair{"cpu_user", rusage.Utime.Usec})
}

func (i *info) dumpMem(buf *bytes.Buffer) {
	buf.WriteString("# Mem\r\n")

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	i.dumpPairs(buf, infoPair{"mem_alloc", mem.Alloc},
		infoPair{"mem_alloc_human", getMemoryHuman(mem.Alloc)})
}

func (i *info) dumpGoroutine(buf *bytes.Buffer) {
	buf.WriteString("# Goroutine\r\n")

	i.dumpPairs(buf, infoPair{"goroutine_num", runtime.NumGoroutine()})
}

func (i *info) dumpPersistence(buf *bytes.Buffer) {
	buf.WriteString("# Persistence\r\n")

	i.dumpPairs(buf, infoPair{"db_name", i.Persistence.DBName})
}

func (i *info) dumpPairs(buf *bytes.Buffer, pairs ...infoPair) {
	for _, v := range pairs {
		buf.WriteString(fmt.Sprintf("%s:%v\r\n", v.Key, v.Value))
	}
}
