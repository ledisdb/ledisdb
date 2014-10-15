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

	Replication struct {
		PubLogNum       int64
		PubLogTotalTime int64 //milliseconds
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
	case "mem":
		i.dumpMem(buf)
	case "store":
		i.dumpStore(buf)
	case "replication":
		i.dumpReplication(buf)
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
	i.dumpStore(buf)
	buf.Write(Delims)
	i.dumpClients(buf)
	buf.Write(Delims)
	i.dumpMem(buf)
	buf.Write(Delims)
	i.dumpReplication(buf)
}

func (i *info) dumpServer(buf *bytes.Buffer) {
	buf.WriteString("# Server\r\n")

	i.dumpPairs(buf, infoPair{"os", i.Server.OS},
		infoPair{"process_id", i.Server.ProceessId},
		infoPair{"addr", i.app.cfg.Addr},
		infoPair{"http_addr", i.app.cfg.HttpAddr},
		infoPair{"readonly", i.app.cfg.Readonly},
		infoPair{"goroutine_num", runtime.NumGoroutine()},
	)
}

func (i *info) dumpClients(buf *bytes.Buffer) {
	buf.WriteString("# Client\r\n")

	i.dumpPairs(buf, infoPair{"client_num", i.Clients.ConnectedClients})
}

func (i *info) dumpMem(buf *bytes.Buffer) {
	buf.WriteString("# Mem\r\n")

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	i.dumpPairs(buf, infoPair{"mem_alloc", mem.Alloc},
		infoPair{"mem_alloc_human", getMemoryHuman(mem.Alloc)})
}

func (i *info) dumpStore(buf *bytes.Buffer) {
	buf.WriteString("# Store\r\n")

	s := i.app.ldb.StoreStat()

	i.dumpPairs(buf, infoPair{"name", i.Persistence.DBName},
		infoPair{"get", s.GetNum},
		infoPair{"get_missing", s.GetMissingNum},
		infoPair{"put", s.PutNum},
		infoPair{"delete", s.DeleteNum},
		infoPair{"iter", s.IterNum},
		infoPair{"iter_seek", s.IterSeekNum},
		infoPair{"iter_close", s.IterCloseNum},
		infoPair{"batch_commit", s.BatchCommitNum},
	)
}

func (i *info) dumpReplication(buf *bytes.Buffer) {
	buf.WriteString("# Replication\r\n")

	p := []infoPair{}
	slaves := make([]string, 0, len(i.app.slaves))
	for s, _ := range i.app.slaves {
		slaves = append(slaves, s.remoteAddr)
	}

	num := i.Replication.PubLogNum
	p = append(p, infoPair{"pub_log_num", num})

	if num != 0 {
		p = append(p, infoPair{"pub_log_per_time", i.Replication.PubLogTotalTime / num})
	} else {
		p = append(p, infoPair{"pub_log_per_time", 0})
	}

	p = append(p, infoPair{"slaveof", i.app.cfg.SlaveOf})

	if len(slaves) > 0 {
		p = append(p, infoPair{"slaves", strings.Join(slaves, ",")})
	}

	if s, _ := i.app.ldb.ReplicationStat(); s != nil {
		p = append(p, infoPair{"last_log_id", s.LastID})
		p = append(p, infoPair{"first_log_id", s.FirstID})
		p = append(p, infoPair{"commit_log_id", s.CommitID})
	} else {
		p = append(p, infoPair{"last_log_id", 0})
		p = append(p, infoPair{"first_log_id", 0})
		p = append(p, infoPair{"commit_log_id", 0})
	}

	i.dumpPairs(buf, p...)
}

func (i *info) dumpPairs(buf *bytes.Buffer, pairs ...infoPair) {
	for _, v := range pairs {
		buf.WriteString(fmt.Sprintf("%s:%v\r\n", v.Key, v.Value))
	}
}
