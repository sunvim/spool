package spool

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// channelPool implements the Pool interface based on buffered channels.
type channelPool struct {
	// storage for our net.Conn connections
	mu    sync.Mutex
	conns chan net.Conn
	// net.Conn generator
	connPool ConnPool
	//had created pool num
	hadCreatedPool int
	//create max pool num
	maxPoolNum int
	//create min pool num
	minPoolNum int
	//check num
	checkTotal int
	//in active conn num
	activeNum int
}

// Factory is a function to create new connections.
type ConnPool func() (net.Conn, error)

// NewChannelPool returns a new pool based on buffered channels with an initial
// capacity and maximum capacity. Factory is used when initial capacity is
// greater than zero to fill the pool. A zero initialCap doesn't fill the Pool
// until a new Get() is called. During a Get(), If there is no new connection
// available in the pool, a new connection will be created via the Factory()
// method.
func NewChannelPool(initialCap, maxCap int, connPool ConnPool) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &channelPool{
		conns:          make(chan net.Conn, maxCap),
		connPool:       connPool,
		hadCreatedPool: 0,
		maxPoolNum:     maxCap,
		minPoolNum:     initialCap,
		checkTotal:     0,
		activeNum:      0,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		conn, err := connPool()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("conn is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}

	//if all the thing is reay,  start the routine of check and compress
	go c.checkAndCompress()
	return c, nil
}

// 5 time/min execute clear pool num
func (c *channelPool) checkAndCompress() {
	tick := time.Tick(30 * time.Second)
	for {
		select {
		case <-tick:
			func() {
				//if had created pool num gt initCap  num , then check time + 1
				if c.hadCreatedPool > 0 {
					c.checkTotal += 1
					//if check times gt 10, then decrease pool num
					if c.checkTotal > 10 {
						conns := c.getConns()
						closeNum := 0
						for closeNum < c.hadCreatedPool {
							conn := <-conns
							conn.Close()
							closeNum += 1
						}
						//reinit value
						c.checkTotal = 0
						c.hadCreatedPool = 0
					}
				}
			}()
		}
	}
}

func (c *channelPool) getConns() chan net.Conn {
	c.mu.Lock()
	conns := c.conns
	c.mu.Unlock()
	return conns
}

// Get implements the Pool interfaces Get() method. If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (c *channelPool) Get() (net.Conn, error) {
	conns := c.getConns()
	if conns == nil {
		return nil, ErrClosed
	}

	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	// if pool num gt max pool, it can not be created any more
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}
		c.activeNum += 1
		return c.wrapConn(conn), nil
	default:
		if c.hadCreatedPool+c.minPoolNum <= c.maxPoolNum {
			conn, err := c.connPool()
			if err != nil {
				return nil, err
			}
			c.hadCreatedPool += 1
			c.activeNum += 1
			return c.wrapConn(conn), nil
		} else {
			return nil, ErrMax
		}
	}
}

// put puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *channelPool) put(conn net.Conn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conns == nil {
		// pool is closed, close passed connection
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.conns <- conn:
		c.activeNum -= 1
		return nil
	default:
		// pool is full, close passed connection
		c.activeNum -= 1
		return conn.Close()
	}
}

func (c *channelPool) Close() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.connPool = nil
	c.mu.Unlock()
	if conns == nil {
		return
	}
	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (c *channelPool) Len() int { return len(c.getConns()) }
