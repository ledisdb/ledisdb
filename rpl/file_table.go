package rpl

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/edsrzf/mmap-go"
	"github.com/siddontang/go/log"
	"github.com/siddontang/go/num"
	"github.com/siddontang/go/sync2"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sync"
	"time"
)

var (
	magic             = []byte("\x1c\x1d\xb8\x88\xff\x9e\x45\x55\x40\xf0\x4c\xda\xe0\xce\x47\xde\x65\x48\x71\x17")
	log0              = Log{0, 1, 1, []byte("ledisdb")}
	log0Data          = []byte{}
	errTableNeedFlush = errors.New("write table need flush")
	errNilHandler     = errors.New("nil write handler")
	pageSize          = int64(4096)
)

func init() {
	log0Data, _ = log0.Marshal()
	pageSize = int64(os.Getpagesize())
}

const tableReaderKeepaliveInterval int64 = 30

func fmtTableName(index int64) string {
	return fmt.Sprintf("%08d.ldb", index)
}

type tableReader struct {
	sync.Mutex

	name  string
	index int64

	f *os.File
	m mmap.MMap

	pf *os.File

	first uint64
	last  uint64

	lastTime uint32

	offsetStartPos int64
	offsetLen      uint32

	lastReadTime sync2.AtomicInt64
}

func newTableReader(base string, index int64) (*tableReader, error) {
	if index <= 0 {
		return nil, fmt.Errorf("invalid index %d", index)
	}
	t := new(tableReader)
	t.name = path.Join(base, fmtTableName(index))
	t.index = index

	var err error

	if err = t.check(); err != nil {
		log.Error("check %s error: %s, try to repair", t.name, err.Error())

		if err = t.repair(); err != nil {
			log.Error("repair %s error: %s", t.name, err.Error())
			return nil, err
		}
	}

	t.close()

	return t, nil
}

func (t *tableReader) Close() {
	t.Lock()
	defer t.Unlock()

	t.close()
}

func (t *tableReader) close() {
	if t.m != nil {
		t.m.Unmap()
		t.m = nil
	}

	if t.f != nil {
		t.f.Close()
		t.f = nil
	}
}

func (t *tableReader) Keepalived() bool {
	l := t.lastReadTime.Get()
	if l > 0 && time.Now().Unix()-l > tableReaderKeepaliveInterval {
		return false
	}

	return true
}

func (t *tableReader) getLogPos(index int) (uint32, error) {
	// if _, err := t.pf.Seek(t.offsetStartPos+int64(index*4), os.SEEK_SET); err != nil {
	// 	return 0, err
	// }

	// var pos uint32
	// if err := binary.Read(t.pf, binary.BigEndian, &pos); err != nil {
	// 	return 0, err
	// }
	// return pos, nil

	return binary.BigEndian.Uint32(t.m[index*4:]), nil
}

func (t *tableReader) check() error {
	var err error

	if t.f, err = os.Open(t.name); err != nil {
		return err
	}

	st, _ := t.f.Stat()

	if st.Size() < 32 {
		return fmt.Errorf("file size %d too short", st.Size())
	}

	var pos int64
	if pos, err = t.f.Seek(-32, os.SEEK_END); err != nil {
		return err
	}

	if err = binary.Read(t.f, binary.BigEndian, &t.offsetStartPos); err != nil {
		return err
	} else if t.offsetStartPos >= st.Size() {
		return fmt.Errorf("invalid offset start pos %d, file size %d", t.offsetStartPos, st.Size())
	} else if t.offsetStartPos%pageSize != 0 {
		return fmt.Errorf("invalid offset start pos %d, must page size %d multi", t.offsetStartPos, pageSize)
	}

	if err = binary.Read(t.f, binary.BigEndian, &t.offsetLen); err != nil {
		return err
	} else if int64(t.offsetLen) >= st.Size() || t.offsetLen == 0 {
		return fmt.Errorf("invalid offset len %d, file size %d", t.offsetLen, st.Size())
	} else if t.offsetLen%4 != 0 {
		return fmt.Errorf("invalid offset len %d, must 4 multiple", t.offsetLen)
	}

	if t.offsetStartPos+int64(t.offsetLen) != pos {
		return fmt.Errorf("invalid offset %d %d", t.offsetStartPos, t.offsetLen)
	}

	b := make([]byte, 20)
	if _, err = t.f.Read(b); err != nil {
		return err
	} else if !bytes.Equal(b, magic) {
		return fmt.Errorf("invalid magic data %q", b)
	}

	if t.m, err = mmap.MapRegion(t.f, int(t.offsetLen), mmap.RDONLY, 0, t.offsetStartPos); err != nil {
		return err
	}

	firstLogPos, _ := t.getLogPos(0)
	lastLogPos, _ := t.getLogPos(int(t.offsetLen/4 - 1))

	if firstLogPos != 0 {
		return fmt.Errorf("invalid first log pos %d, must 0", firstLogPos)
	} else if int64(lastLogPos) > t.offsetStartPos {
		return fmt.Errorf("invalid last log pos %d", lastLogPos)
	}

	var l Log
	if _, err = t.decodeLogHead(&l, int64(firstLogPos)); err != nil {
		return fmt.Errorf("decode first log err %s", err.Error())
	}

	t.first = l.ID
	var n int64
	if n, err = t.decodeLogHead(&l, int64(lastLogPos)); err != nil {
		return fmt.Errorf("decode last log err %s", err.Error())
	} else {
		var l0 Log
		if _, err := t.f.Seek(n, os.SEEK_SET); err != nil {
			return fmt.Errorf("seek logo err %s", err.Error())
		} else if err = l0.Decode(t.f); err != nil {
			println(lastLogPos, n, l0.ID, l0.CreateTime, l0.Compression)
			return fmt.Errorf("decode log0 err %s", err.Error())
		} else if !reflect.DeepEqual(l0, log0) {
			return fmt.Errorf("invalid log0 %#v != %#v", l0, log0)
		}
	}

	t.last = l.ID
	t.lastTime = l.CreateTime

	if t.first > t.last {
		return fmt.Errorf("invalid log table first %d > last %d", t.first, t.last)
	} else if (t.last - t.first + 1) != uint64(t.offsetLen/4) {
		return fmt.Errorf("invalid log table, first %d, last %d, and log num %d", t.first, t.last, t.offsetLen/4)
	}

	return nil
}

func (t *tableReader) repair() error {
	t.close()

	var err error
	if t.f, err = os.Open(t.name); err != nil {
		return err
	}

	defer t.close()

	st, _ := t.f.Stat()
	size := st.Size()

	if size == 0 {
		return fmt.Errorf("empty file, can not repaired")
	}

	tw := newTableWriter(path.Dir(t.name), t.index, maxLogFileSize)

	tmpName := tw.name + ".tmp"
	tw.name = tmpName
	os.Remove(tmpName)

	defer func() {
		tw.Close()
		os.Remove(tmpName)
	}()

	var l Log

	for {
		lastPos, _ := t.f.Seek(0, os.SEEK_CUR)
		if lastPos == size {
			//no data anymore, we can not read log0
			//we may meet the log missing risk but have no way
			log.Error("no more data, maybe missing some logs, use your own risk!!!")
			break
		}

		if err := l.Decode(t.f); err != nil {
			return err
		}

		if l.ID == 0 {
			break
		}

		t.lastTime = l.CreateTime

		if err := tw.StoreLog(&l); err != nil {
			return err
		}
	}

	t.close()

	var tr *tableReader
	if tr, err = tw.Flush(); err != nil {
		return err
	}

	t.first = tr.first
	t.last = tr.last
	t.offsetStartPos = tr.offsetStartPos
	t.offsetLen = tr.offsetLen

	defer tr.Close()

	os.Remove(t.name)

	if err := os.Rename(tmpName, t.name); err != nil {
		return err
	}

	return nil
}

func (t *tableReader) decodeLogHead(l *Log, pos int64) (int64, error) {
	_, err := t.f.Seek(int64(pos), os.SEEK_SET)
	if err != nil {
		return 0, err
	}

	dataLen, err := l.DecodeHead(t.f)
	if err != nil {
		return 0, err
	}

	return pos + int64(l.HeadSize()) + int64(dataLen), nil
}

func (t *tableReader) GetLog(id uint64, l *Log) error {
	if id < t.first || id > t.last {
		return ErrLogNotFound
	}

	t.lastReadTime.Set(time.Now().Unix())

	t.Lock()
	defer t.Unlock()

	if err := t.openTable(); err != nil {
		t.close()
		return err
	}

	pos, err := t.getLogPos(int(id - t.first))
	if err != nil {
		return err
	}

	if _, err := t.f.Seek(int64(pos), os.SEEK_SET); err != nil {
		return err
	}

	if err := l.Decode(t.f); err != nil {
		return err
	} else if l.ID != id {
		return fmt.Errorf("invalid log id %d != %d", l.ID, id)
	}

	return nil
}

func (t *tableReader) openTable() error {
	var err error
	if t.f == nil {
		if t.f, err = os.Open(t.name); err != nil {
			return err
		}
	}

	if t.m == nil {
		if t.m, err = mmap.MapRegion(t.f, int(t.offsetLen), mmap.RDONLY, 0, t.offsetStartPos); err != nil {
			return err
		}
	}

	return nil
}

type tableWriter struct {
	sync.RWMutex

	wf *os.File
	rf *os.File

	wb *bufio.Writer

	rm sync.Mutex

	base  string
	name  string
	index int64

	first uint64
	last  uint64

	offsetPos int64
	offsetBuf []byte

	maxLogSize int64

	closed bool

	syncType int
	lastTime uint32

	// cache *logLRUCache
}

func newTableWriter(base string, index int64, maxLogSize int64) *tableWriter {
	if index <= 0 {
		panic(fmt.Errorf("invalid index %d", index))
	}

	t := new(tableWriter)

	t.base = base
	t.name = path.Join(base, fmtTableName(index))
	t.index = index

	t.offsetPos = 0
	t.maxLogSize = maxLogSize

	//maybe config later?
	t.wb = bufio.NewWriterSize(ioutil.Discard, 4096)

	t.closed = false

	//maybe use config later
	// t.cache = newLogLRUCache(1024*1024, 1000)

	return t
}

func (t *tableWriter) SetMaxLogSize(s int64) {
	t.maxLogSize = s
}

func (t *tableWriter) SetSyncType(tp int) {
	t.syncType = tp
}

func (t *tableWriter) close() {
	if t.rf != nil {
		t.rf.Close()
		t.rf = nil
	}

	if t.wf != nil {
		t.wf.Close()
		t.wf = nil
	}

	t.wb.Reset(ioutil.Discard)
}

func (t *tableWriter) Close() {
	t.Lock()
	defer t.Unlock()

	t.closed = true

	t.close()
}

func (t *tableWriter) First() uint64 {
	t.Lock()
	id := t.first
	t.Unlock()
	return id
}

func (t *tableWriter) Last() uint64 {
	t.Lock()
	id := t.last
	t.Unlock()
	return id
}

func (t *tableWriter) reset() {
	t.close()

	t.first = 0
	t.last = 0
	t.index = t.index + 1
	t.name = path.Join(t.base, fmtTableName(t.index))
	t.offsetBuf = t.offsetBuf[0:0]
	t.offsetPos = 0
	// t.cache.Reset()
}

func (t *tableWriter) Flush() (*tableReader, error) {
	t.Lock()
	defer t.Unlock()

	if t.wf == nil {
		return nil, errNilHandler
	}

	defer t.reset()

	tr := new(tableReader)
	tr.name = t.name
	tr.index = t.index

	st, _ := t.wf.Stat()

	tr.first = t.first
	tr.last = t.last

	if n, err := t.wf.Write(log0Data); err != nil {
		return nil, fmt.Errorf("flush log0data error %s", err.Error())
	} else if n != len(log0Data) {
		return nil, fmt.Errorf("flush log0data only %d != %d", n, len(log0Data))
	}

	st, _ = t.wf.Stat()

	if m := st.Size() % pageSize; m != 0 {
		padding := pageSize - m
		if n, err := t.wf.Write(make([]byte, padding)); err != nil {
			return nil, fmt.Errorf("flush log padding error %s", err.Error())
		} else if n != int(padding) {
			return nil, fmt.Errorf("flush log padding error %d != %d", n, padding)
		}
	}

	st, _ = t.wf.Stat()

	if st.Size()%pageSize != 0 {
		return nil, fmt.Errorf("invalid offset start pos, %d", st.Size())
	}

	tr.offsetStartPos = st.Size()
	tr.offsetLen = uint32(len(t.offsetBuf))

	if n, err := t.wf.Write(t.offsetBuf); err != nil {
		log.Error("flush offset buffer error %s", err.Error())
		return nil, err
	} else if n != len(t.offsetBuf) {
		log.Error("flush offset buffer only %d != %d", n, len(t.offsetBuf))
		return nil, io.ErrShortWrite
	}

	if err := binary.Write(t.wf, binary.BigEndian, tr.offsetStartPos); err != nil {
		log.Error("flush offset start pos error %s", err.Error())
		return nil, err
	}

	if err := binary.Write(t.wf, binary.BigEndian, tr.offsetLen); err != nil {
		log.Error("flush offset len error %s", err.Error())
		return nil, err
	}

	if n, err := t.wf.Write(magic); err != nil {
		log.Error("flush magic data error %s", err.Error())
		return nil, err
	} else if n != len(magic) {
		log.Error("flush magic data only %d != %d", n, len(magic))
		return nil, io.ErrShortWrite
	}

	return tr, nil
}

func (t *tableWriter) StoreLog(l *Log) error {
	if l.ID == 0 {
		return ErrStoreLogID
	}

	t.Lock()
	defer t.Unlock()

	if t.closed {
		return fmt.Errorf("table writer is closed")
	}

	if t.last > 0 && l.ID != t.last+1 {
		return ErrStoreLogID
	}

	if t.last-t.first+1 > maxLogNumInFile {
		return errTableNeedFlush
	}

	var err error
	if t.wf == nil {
		if t.wf, err = os.OpenFile(t.name, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return err
		}
		t.wb.Reset(t.wf)
	}

	if t.offsetBuf == nil {
		t.offsetBuf = make([]byte, 0, maxLogNumInFile*4)
	}

	// st, _ := t.wf.Stat()
	// if st.Size() >= t.maxLogSize {
	// 	return errTableNeedFlush
	// }

	if t.offsetPos >= t.maxLogSize {
		return errTableNeedFlush
	}

	offsetPos := t.offsetPos

	if err := l.Encode(t.wb); err != nil {
		return err
	} else if err = t.wb.Flush(); err != nil {
		return err
	}

	// buf, _ := l.Marshal()
	// if n, err := t.wf.Write(buf); err != nil {
	// 	return err
	// } else if n != len(buf) {
	// 	return io.ErrShortWrite
	// }

	t.offsetPos += int64(l.Size())

	t.offsetBuf = append(t.offsetBuf, num.Uint32ToBytes(uint32(offsetPos))...)
	if t.first == 0 {
		t.first = l.ID
	}

	t.last = l.ID

	t.lastTime = l.CreateTime

	// t.cache.Set(l.ID, buf)

	if t.syncType == 2 {
		if err := t.wf.Sync(); err != nil {
			log.Error("sync table error %s", err.Error())
		}
	}

	return nil
}

func (t *tableWriter) GetLog(id uint64, l *Log) error {
	t.RLock()
	defer t.RUnlock()

	if id < t.first || id > t.last {
		return ErrLogNotFound
	}

	// if cl := t.cache.Get(id); cl != nil {
	// 	if err := l.Unmarshal(cl); err == nil && l.ID == id {
	// 		return nil
	// 	} else {
	// 		t.cache.Delete(id)
	// 	}
	// }

	offset := binary.BigEndian.Uint32(t.offsetBuf[(id-t.first)*4:])

	if err := t.getLog(l, int64(offset)); err != nil {
		return err
	} else if l.ID != id {
		return fmt.Errorf("invalid log id %d != %d", id, l.ID)
	}

	return nil
}

func (t *tableWriter) Sync() error {
	t.Lock()
	defer t.Unlock()

	if t.wf != nil {
		return t.wf.Sync()
	}
	return nil
}

func (t *tableWriter) getLog(l *Log, pos int64) error {
	t.rm.Lock()
	defer t.rm.Unlock()

	var err error
	if t.rf == nil {
		if t.rf, err = os.Open(t.name); err != nil {
			return err
		}
	}

	if _, err = t.rf.Seek(pos, os.SEEK_SET); err != nil {
		return err
	}

	if err = l.Decode(t.rf); err != nil {
		return err
	}

	return nil
}
