package server

import (
	"bufio"
	"errors"
	"github.com/siddontang/go/arena"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/log"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"net"
	"runtime"
	"strconv"
	"time"
)

var errReadRequest = errors.New("invalid request protocol")
var errClientQuit = errors.New("remote client quit")

type respClient struct {
	*client

	conn net.Conn
	rb   *bufio.Reader

	ar *arena.Arena

	activeQuit bool
}

type respWriter struct {
	buff *bufio.Writer
}

func (app *App) addRespClient(c *respClient) {
	app.rcm.Lock()
	app.rcs[c] = struct{}{}
	app.rcm.Unlock()
}

func (app *App) delRespClient(c *respClient) {
	app.rcm.Lock()
	delete(app.rcs, c)
	app.rcm.Unlock()
}

func (app *App) closeAllRespClients() {
	app.rcm.Lock()

	for c := range app.rcs {
		c.conn.Close()
	}

	app.rcm.Unlock()
}

func (app *App) respClientNum() int {
	app.rcm.Lock()
	n := len(app.rcs)
	app.rcm.Unlock()
	return n
}

func newClientRESP(conn net.Conn, app *App) {
	c := new(respClient)

	c.client = newClient(app)
	c.conn = conn

	c.activeQuit = false

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetReadBuffer(app.cfg.ConnReadBufferSize)
		tcpConn.SetWriteBuffer(app.cfg.ConnWriteBufferSize)
	}

	c.rb = bufio.NewReaderSize(conn, app.cfg.ConnReadBufferSize)

	c.resp = newWriterRESP(conn, app.cfg.ConnWriteBufferSize)
	c.remoteAddr = conn.RemoteAddr().String()

	//maybe another config?
	c.ar = arena.NewArena(app.cfg.ConnReadBufferSize)

	app.connWait.Add(1)

	app.addRespClient(c)

	go c.run()
}

func (c *respClient) run() {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			buf = buf[0:n]

			log.Fatal("client run panic %s:%v", buf, e)
		}

		c.client.close()

		c.conn.Close()

		if c.tx != nil {
			c.tx.Rollback()
			c.tx = nil
		}

		c.app.removeSlave(c.client, c.activeQuit)

		c.app.delRespClient(c)

		c.app.connWait.Done()
	}()

	select {
	case <-c.app.quit:
		//check app closed
		return
	default:
		break
	}

	kc := time.Duration(c.app.cfg.ConnKeepaliveInterval) * time.Second
	for {
		if kc > 0 {
			c.conn.SetReadDeadline(time.Now().Add(kc))
		}

		reqData, err := c.readRequest()
		if err == nil {
			err = c.handleRequest(reqData)
		}

		if err != nil {
			return
		}
	}
}

func (c *respClient) readRequest() ([][]byte, error) {
	return ReadRequest(c.rb, c.ar)
}

func (c *respClient) handleRequest(reqData [][]byte) error {
	if len(reqData) == 0 {
		c.cmd = ""
		c.args = reqData[0:0]
	} else {
		c.cmd = hack.String(lowerSlice(reqData[0]))
		c.args = reqData[1:]
	}
	if c.cmd == "quit" {
		c.activeQuit = true
		c.resp.writeStatus(OK)
		c.resp.flush()
		c.conn.Close()
		return errClientQuit
	}

	c.perform()

	c.cmd = ""
	c.args = nil

	c.ar.Reset()

	return nil
}

//	response writer

func newWriterRESP(conn net.Conn, size int) *respWriter {
	w := new(respWriter)
	w.buff = bufio.NewWriterSize(conn, size)
	return w
}

func (w *respWriter) writeError(err error) {
	w.buff.Write(hack.Slice("-ERR"))
	if err != nil {
		w.buff.WriteByte(' ')
		w.buff.Write(hack.Slice(err.Error()))
	}
	w.buff.Write(Delims)
}

func (w *respWriter) writeStatus(status string) {
	w.buff.WriteByte('+')
	w.buff.Write(hack.Slice(status))
	w.buff.Write(Delims)
}

func (w *respWriter) writeInteger(n int64) {
	w.buff.WriteByte(':')
	w.buff.Write(num.FormatInt64ToSlice(n))
	w.buff.Write(Delims)
}

func (w *respWriter) writeBulk(b []byte) {
	w.buff.WriteByte('$')
	if b == nil {
		w.buff.Write(NullBulk)
	} else {
		w.buff.Write(hack.Slice(strconv.Itoa(len(b))))
		w.buff.Write(Delims)
		w.buff.Write(b)
	}

	w.buff.Write(Delims)
}

func (w *respWriter) writeArray(lst []interface{}) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(NullArray)
		w.buff.Write(Delims)
	} else {
		w.buff.Write(hack.Slice(strconv.Itoa(len(lst))))
		w.buff.Write(Delims)

		for i := 0; i < len(lst); i++ {
			switch v := lst[i].(type) {
			case []interface{}:
				w.writeArray(v)
			case [][]byte:
				w.writeSliceArray(v)
			case []byte:
				w.writeBulk(v)
			case nil:
				w.writeBulk(nil)
			case int64:
				w.writeInteger(v)
			default:
				panic("invalid array type")
			}
		}
	}
}

func (w *respWriter) writeSliceArray(lst [][]byte) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(NullArray)
		w.buff.Write(Delims)
	} else {
		w.buff.Write(hack.Slice(strconv.Itoa(len(lst))))
		w.buff.Write(Delims)

		for i := 0; i < len(lst); i++ {
			w.writeBulk(lst[i])
		}
	}
}

func (w *respWriter) writeFVPairArray(lst []ledis.FVPair) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(NullArray)
		w.buff.Write(Delims)
	} else {
		w.buff.Write(hack.Slice(strconv.Itoa(len(lst) * 2)))
		w.buff.Write(Delims)

		for i := 0; i < len(lst); i++ {
			w.writeBulk(lst[i].Field)
			w.writeBulk(lst[i].Value)
		}
	}
}

func (w *respWriter) writeScorePairArray(lst []ledis.ScorePair, withScores bool) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(NullArray)
		w.buff.Write(Delims)
	} else {
		if withScores {
			w.buff.Write(hack.Slice(strconv.Itoa(len(lst) * 2)))
			w.buff.Write(Delims)
		} else {
			w.buff.Write(hack.Slice(strconv.Itoa(len(lst))))
			w.buff.Write(Delims)

		}

		for i := 0; i < len(lst); i++ {
			w.writeBulk(lst[i].Member)

			if withScores {
				w.writeBulk(num.FormatInt64ToSlice(lst[i].Score))
			}
		}
	}
}

func (w *respWriter) writeBulkFrom(n int64, rb io.Reader) {
	w.buff.WriteByte('$')
	w.buff.Write(hack.Slice(strconv.FormatInt(n, 10)))
	w.buff.Write(Delims)

	io.Copy(w.buff, rb)
	w.buff.Write(Delims)
}

func (w *respWriter) flush() {
	w.buff.Flush()
}
