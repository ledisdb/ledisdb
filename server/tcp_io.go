package server

import (
	"bufio"
	"errors"
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"net"
	"strconv"
)

type tcpContext struct {
	conn net.Conn
}

type tcpWriter struct {
	buff *bufio.Writer
}

type tcpReader struct {
	buff *bufio.Reader
}

//	tcp context

func newTcpContext(conn net.Conn) *tcpContext {
	ctx := new(tcpContext)
	ctx.conn = conn
	return ctx
}

func (ctx *tcpContext) addr() string {
	return ctx.conn.RemoteAddr().String()
}

func (ctx *tcpContext) release() {
	if ctx.conn != nil {
		ctx.conn.Close()
		ctx.conn = nil
	}
}

//	tcp reader

func newTcpReader(conn net.Conn) *tcpReader {
	r := new(tcpReader)
	r.buff = bufio.NewReaderSize(conn, 256)
	return r
}

func (r *tcpReader) readLine() ([]byte, error) {
	return ReadLine(r.buff)
}

//A client sends to the Redis server a RESP Array consisting of just Bulk Strings.
func (r *tcpReader) read() ([][]byte, error) {
	l, err := r.readLine()
	if err != nil {
		return nil, err
	} else if len(l) == 0 || l[0] != '*' {
		return nil, errReadRequest
	}

	var nparams int
	if nparams, err = strconv.Atoi(ledis.String(l[1:])); err != nil {
		return nil, err
	} else if nparams <= 0 {
		return nil, errReadRequest
	}

	reqData := make([][]byte, 0, nparams)
	var n int
	for i := 0; i < nparams; i++ {
		if l, err = r.readLine(); err != nil {
			return nil, err
		}

		if len(l) == 0 {
			return nil, errReadRequest
		} else if l[0] == '$' {
			//handle resp string
			if n, err = strconv.Atoi(ledis.String(l[1:])); err != nil {
				return nil, err
			} else if n == -1 {
				reqData = append(reqData, nil)
			} else {
				buf := make([]byte, n)
				if _, err = io.ReadFull(r.buff, buf); err != nil {
					return nil, err
				}

				if l, err = r.readLine(); err != nil {
					return nil, err
				} else if len(l) != 0 {
					return nil, errors.New("bad bulk string format")
				}

				reqData = append(reqData, buf)

			}

		} else {
			return nil, errReadRequest
		}
	}

	return reqData, nil
}

//	tcp writer

func newTcpWriter(conn net.Conn) *tcpWriter {
	w := new(tcpWriter)
	w.buff = bufio.NewWriterSize(conn, 256)
	return w
}

func (w *tcpWriter) writeError(err error) {
	w.buff.Write(ledis.Slice("-ERR"))
	if err != nil {
		w.buff.WriteByte(' ')
		w.buff.Write(ledis.Slice(err.Error()))
	}
	w.buff.Write(Delims)
}

func (w *tcpWriter) writeStatus(status string) {
	w.buff.WriteByte('+')
	w.buff.Write(ledis.Slice(status))
	w.buff.Write(Delims)
}

func (w *tcpWriter) writeInteger(n int64) {
	w.buff.WriteByte(':')
	w.buff.Write(ledis.StrPutInt64(n))
	w.buff.Write(Delims)
}

func (w *tcpWriter) writeBulk(b []byte) {
	w.buff.WriteByte('$')
	if b == nil {
		w.buff.Write(NullBulk)
	} else {
		w.buff.Write(ledis.Slice(strconv.Itoa(len(b))))
		w.buff.Write(Delims)
		w.buff.Write(b)
	}

	w.buff.Write(Delims)
}

func (w *tcpWriter) writeArray(lst []interface{}) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(NullArray)
		w.buff.Write(Delims)
	} else {
		w.buff.Write(ledis.Slice(strconv.Itoa(len(lst))))
		w.buff.Write(Delims)

		for i := 0; i < len(lst); i++ {
			switch v := lst[i].(type) {
			case []interface{}:
				w.writeArray(v)
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

func (w *tcpWriter) writeSliceArray(lst [][]byte) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(NullArray)
		w.buff.Write(Delims)
	} else {
		w.buff.Write(ledis.Slice(strconv.Itoa(len(lst))))
		w.buff.Write(Delims)

		for i := 0; i < len(lst); i++ {
			w.writeBulk(lst[i])
		}
	}
}

func (w *tcpWriter) writeFVPairArray(lst []ledis.FVPair) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(NullArray)
		w.buff.Write(Delims)
	} else {
		w.buff.Write(ledis.Slice(strconv.Itoa(len(lst) * 2)))
		w.buff.Write(Delims)

		for i := 0; i < len(lst); i++ {
			w.writeBulk(lst[i].Field)
			w.writeBulk(lst[i].Value)
		}
	}
}

func (w *tcpWriter) writeScorePairArray(lst []ledis.ScorePair, withScores bool) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(NullArray)
		w.buff.Write(Delims)
	} else {
		if withScores {
			w.buff.Write(ledis.Slice(strconv.Itoa(len(lst) * 2)))
			w.buff.Write(Delims)
		} else {
			w.buff.Write(ledis.Slice(strconv.Itoa(len(lst))))
			w.buff.Write(Delims)

		}

		for i := 0; i < len(lst); i++ {
			w.writeBulk(lst[i].Member)

			if withScores {
				w.writeBulk(ledis.StrPutInt64(lst[i].Score))
			}
		}
	}
}

func (w *tcpWriter) writeBulkFrom(n int64, rb io.Reader) {
	w.buff.WriteByte('$')
	w.buff.Write(ledis.Slice(strconv.FormatInt(n, 10)))
	w.buff.Write(Delims)

	io.Copy(w.buff, rb)
	w.buff.Write(Delims)
}

func (w *tcpWriter) flush() {
	w.buff.Flush()
}
