package server

import (
	"bytes"
	"fmt"
	"github.com/siddontang/go/sync2"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type info struct {
	sync.Mutex

	app *App

	Server struct {
		OS         string
		ProceessId int
	}

	Replication struct {
		PubLogNum          sync2.AtomicInt64
		PubLogAckNum       sync2.AtomicInt64
		PubLogTotalAckTime sync2.AtomicDuration

		MasterLastLogID sync2.AtomicUint64
	}
}

func newInfo(app *App) (i *info, err error) {
	i = new(info)

	i.app = app

	i.Server.OS = runtime.GOOS
	i.Server.ProceessId = os.Getpid()

	return i, nil
}

func (i *info) Close() {

}

func getMemoryHuman(m uint64) string {
	if m > GB {
		return fmt.Sprintf("%0.3fG", float64(m)/float64(GB))
	} else if m > MB {
		return fmt.Sprintf("%0.3fM", float64(m)/float64(MB))
	} else if m > KB {
		return fmt.Sprintf("%0.3fK", float64(m)/float64(KB))
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
	case "mem":
		i.dumpMem(buf)
	case "gc":
		i.dumpGC(buf)
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
	i.dumpMem(buf)
	buf.Write(Delims)
	i.dumpGC(buf)
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
		infoPair{"cgo_call_num", runtime.NumCgoCall()},
		infoPair{"resp_client_num", i.app.respClientNum()},
	)
}

func (i *info) dumpMem(buf *bytes.Buffer) {
	buf.WriteString("# Mem\r\n")

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	i.dumpPairs(buf, infoPair{"mem_alloc", getMemoryHuman(mem.Alloc)},
		infoPair{"mem_sys", getMemoryHuman(mem.Sys)},
		infoPair{"mem_looksups", getMemoryHuman(mem.Lookups)},
		infoPair{"mem_mallocs", getMemoryHuman(mem.Mallocs)},
		infoPair{"mem_frees", getMemoryHuman(mem.Frees)},
		infoPair{"mem_total", getMemoryHuman(mem.TotalAlloc)},
		infoPair{"mem_heap_alloc", getMemoryHuman(mem.HeapAlloc)},
		infoPair{"mem_heap_sys", getMemoryHuman(mem.HeapSys)},
		infoPair{"mem_head_idle", getMemoryHuman(mem.HeapIdle)},
		infoPair{"mem_head_inuse", getMemoryHuman(mem.HeapInuse)},
		infoPair{"mem_head_released", getMemoryHuman(mem.HeapReleased)},
		infoPair{"mem_head_objects", mem.HeapObjects},
	)
}

const (
	gcTimeFormat = "2006/01/02 15:04:05.000"
)

func (i *info) dumpGC(buf *bytes.Buffer) {
	buf.WriteString("# GC\r\n")

	count := 5

	var st debug.GCStats
	st.Pause = make([]time.Duration, count)
	// st.PauseQuantiles = make([]time.Duration, count)
	debug.ReadGCStats(&st)

	h := make([]string, 0, count)

	for i := 0; i < count && i < len(st.Pause); i++ {
		h = append(h, st.Pause[i].String())
	}

	i.dumpPairs(buf, infoPair{"gc_last_time", st.LastGC.Format(gcTimeFormat)},
		infoPair{"gc_num", st.NumGC},
		infoPair{"gc_pause_total", st.PauseTotal.String()},
		infoPair{"gc_pause_history", strings.Join(h, ",")},
	)
}

func (i *info) dumpStore(buf *bytes.Buffer) {
	buf.WriteString("# Store\r\n")

	s := i.app.ldb.StoreStat()

	// getNum := s.GetNum.Get()
	// getTotalTime := s.GetTotalTime.Get()

	// gt := int64(0)
	// if getNum > 0 {
	// 	gt = getTotalTime.Nanoseconds() / (getNum * 1e3)
	// }

	// commitNum := s.BatchCommitNum.Get()
	// commitTotalTime := s.BatchCommitTotalTime.Get()

	// ct := int64(0)
	// if commitNum > 0 {
	// 	ct = commitTotalTime.Nanoseconds() / (commitNum * 1e3)
	// }

	i.dumpPairs(buf, infoPair{"name", i.app.cfg.DBName},
		infoPair{"get", s.GetNum},
		infoPair{"get_missing", s.GetMissingNum},
		infoPair{"put", s.PutNum},
		infoPair{"delete", s.DeleteNum},
		infoPair{"get_total_time", s.GetTotalTime.Get().String()},
		infoPair{"iter", s.IterNum},
		infoPair{"iter_seek", s.IterSeekNum},
		infoPair{"iter_close", s.IterCloseNum},
		infoPair{"batch_commit", s.BatchCommitNum},
		infoPair{"batch_commit_total_time", s.BatchCommitTotalTime.Get().String()},
	)
}

func (i *info) dumpReplication(buf *bytes.Buffer) {
	buf.WriteString("# Replication\r\n")

	p := []infoPair{}
	i.app.slock.Lock()
	slaves := make([]string, 0, len(i.app.slaves))
	for _, s := range i.app.slaves {
		slaves = append(slaves, s.slaveListeningAddr)
	}
	i.app.slock.Unlock()

	num := i.Replication.PubLogNum.Get()
	p = append(p, infoPair{"pub_log_num", num})

	ackNum := i.Replication.PubLogAckNum.Get()
	totalTime := i.Replication.PubLogTotalAckTime.Get().Nanoseconds() / 1e6
	if ackNum != 0 {
		p = append(p, infoPair{"pub_log_ack_per_time", totalTime / ackNum})
	} else {
		p = append(p, infoPair{"pub_log_ack_per_time", 0})
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

	p = append(p, infoPair{"master_last_log_id", i.Replication.MasterLastLogID.Get()})

	i.dumpPairs(buf, p...)
}

func (i *info) dumpPairs(buf *bytes.Buffer, pairs ...infoPair) {
	for _, v := range pairs {
		buf.WriteString(fmt.Sprintf("%s:%v\r\n", v.Key, v.Value))
	}
}
