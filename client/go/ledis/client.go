package ledis

import (
	"container/list"
	"strings"
	"sync"
	"time"
)

const (
	pingPeriod time.Duration = 3 * time.Second
)

type Config struct {
	Addr         string
	MaxIdleConns int
}

type Client struct {
	sync.Mutex

	cfg   *Config
	proto string

	conns *list.List
}

func NewClient(cfg *Config) *Client {
	c := new(Client)

	c.cfg = cfg

	if strings.Contains(cfg.Addr, "/") {
		c.proto = "unix"
	} else {
		c.proto = "tcp"
	}

	c.conns = list.New()

	return c
}

func (c *Client) Do(cmd string, args ...interface{}) (interface{}, error) {
	co := c.get()
	r, err := co.Do(cmd, args...)
	c.put(co)

	return r, err
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

func (c *Client) Get() *Conn {
	return c.get()
}

func (c *Client) get() *Conn {
	c.Lock()
	if c.conns.Len() == 0 {
		c.Unlock()

		return c.newConn()
	} else {
		e := c.conns.Front()
		co := e.Value.(*Conn)
		c.conns.Remove(e)

		c.Unlock()

		return co
	}
}

func (c *Client) put(conn *Conn) {
	c.Lock()
	if c.conns.Len() >= c.cfg.MaxIdleConns {
		c.Unlock()
		conn.finalize()
	} else {
		conn.lastActive = time.Now()
		c.conns.PushFront(conn)
		c.Unlock()
	}
}
