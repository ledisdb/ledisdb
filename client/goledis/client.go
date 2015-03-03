package ledis

import (
	"container/list"
	"net"
	"strings"
	"sync"
)

type Config struct {
	Addr            string
	MaxIdleConns    int
	ReadBufferSize  int
	WriteBufferSize int
}

type Client struct {
	sync.Mutex

	cfg *Config

	conns *list.List
}

func getProto(addr string) string {
	if strings.Contains(addr, "/") {
		return "unix"
	} else {
		return "tcp"
	}
}

func NewClient(cfg *Config) *Client {
	c := new(Client)

	c.cfg = cfg
	if c.cfg.ReadBufferSize == 0 {
		c.cfg.ReadBufferSize = 4096
	}
	if c.cfg.WriteBufferSize == 0 {
		c.cfg.WriteBufferSize = 4096
	}

	c.conns = list.New()

	return c
}

func (c *Client) Do(cmd string, args ...interface{}) (interface{}, error) {
	var co *Conn
	var err error
	var r interface{}

	for i := 0; i < 2; i++ {
		co, err = c.get()
		if err != nil {
			return nil, err
		}

		r, err = co.Do(cmd, args...)
		if err != nil {
			co.finalize()

			if e, ok := err.(*net.OpError); ok && strings.Contains(e.Error(), "use of closed network connection") {
				//send to a closed connection, try again
				continue
			}

			return nil, err
		} else {
			c.put(co)
		}

		return r, nil
	}

	return nil, err
}

func (c *Client) Close() {
	c.Lock()
	defer c.Unlock()

	for c.conns.Len() > 0 {
		e := c.conns.Front()
		co := e.Value.(*Conn)
		c.conns.Remove(e)

		co.finalize()
	}
}

func (c *Client) Get() (*Conn, error) {
	return c.get()
}

func (c *Client) get() (*Conn, error) {
	c.Lock()
	if c.conns.Len() == 0 {
		c.Unlock()

		return c.newConn(c.cfg.Addr)
	} else {
		e := c.conns.Front()
		co := e.Value.(*Conn)
		c.conns.Remove(e)

		c.Unlock()

		return co, nil
	}
}

func (c *Client) put(conn *Conn) {
	c.Lock()
	if c.conns.Len() >= c.cfg.MaxIdleConns {
		c.Unlock()
		conn.finalize()
	} else {
		c.conns.PushFront(conn)
		c.Unlock()
	}
}
