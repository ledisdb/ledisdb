package goredis

import (
	"container/list"
	"net"
	"strings"
	"sync"
	"time"
)

type PoolConn struct {
	*Conn
	c *Client
}

func (c *PoolConn) Close() {
	if c.Conn.isClosed() {
		return
	}

	c.c.put(c.Conn)
}

// force close inner connection and not put it into pool
func (c *PoolConn) Finalize() {
	c.Conn.Close()
}

type Client struct {
	sync.Mutex

	addr            string
	maxIdleConns    int
	readBufferSize  int
	writeBufferSize int
	password        string

	conns *list.List

	quit chan struct{}
	wg   sync.WaitGroup
}

func getProto(addr string) string {
	if strings.Contains(addr, "/") {
		return "unix"
	} else {
		return "tcp"
	}
}

func NewClient(addr string, password string) *Client {
	c := new(Client)

	c.addr = addr
	c.maxIdleConns = 4
	c.readBufferSize = 1024
	c.writeBufferSize = 1024
	c.password = password

	c.conns = list.New()
	c.quit = make(chan struct{})

	c.wg.Add(1)
	go c.onCheck()

	return c
}

func (c *Client) SetPassword(pass string) {
	c.password = pass
}

func (c *Client) SetReadBufferSize(s int) {
	c.readBufferSize = s
}

func (c *Client) SetWriteBufferSize(s int) {
	c.writeBufferSize = s
}

func (c *Client) SetMaxIdleConns(n int) {
	c.maxIdleConns = n
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
			co.Close()

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

	close(c.quit)
	c.wg.Wait()

	for c.conns.Len() > 0 {
		e := c.conns.Front()
		co := e.Value.(*Conn)
		c.conns.Remove(e)

		co.Close()
	}
}

func (c *Client) Get() (*PoolConn, error) {
	co, err := c.get()
	if err != nil {
		return nil, err
	}

	return &PoolConn{co, c}, err
}

func (c *Client) get() (co *Conn, err error) {
	c.Lock()
	if c.conns.Len() == 0 {
		c.Unlock()

		co, err = c.newConn(c.addr, c.password)
	} else {
		e := c.conns.Front()
		co = e.Value.(*Conn)
		c.conns.Remove(e)

		c.Unlock()
	}

	return
}

func (c *Client) put(conn *Conn) {
	c.Lock()
	defer c.Unlock()

	for c.conns.Len() >= c.maxIdleConns {
		// remove back
		e := c.conns.Back()
		co := e.Value.(*Conn)
		c.conns.Remove(e)

		co.Close()
	}

	c.conns.PushFront(conn)
}

func (c *Client) getIdle() *Conn {
	c.Lock()
	defer c.Unlock()

	if c.conns.Len() == 0 {
		return nil
	} else {
		e := c.conns.Back()
		co := e.Value.(*Conn)
		c.conns.Remove(e)
		return co
	}
}

func (c *Client) checkIdle() {
	co := c.getIdle()
	if co == nil {
		return
	}

	_, err := co.Do("PING")
	if err != nil {
		co.Close()
	} else {
		c.put(co)
	}
}

func (c *Client) onCheck() {
	t := time.NewTicker(3 * time.Second)

	defer func() {
		t.Stop()
		c.wg.Done()
	}()

	for {
		select {
		case <-t.C:
			c.checkIdle()
		case <-c.quit:
			return
		}
	}
}
