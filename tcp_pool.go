package ethernet_ip

import (
	"errors"
	"sync"
)

// EIPTCPPool provides a connection pool for EtherNet/IP TCP connections.
// It is useful for high-performance scenarios where multiple concurrent
// connections are needed.
type EIPTCPPool struct {
	factory func() (*EIPTCP, error)
	pool    chan *EIPTCP
	mu      sync.Mutex
	closed  bool
}

// NewTCPPool creates a new connection pool for the specified address.
// The capacity parameter determines the maximum number of idle connections.
// If capacity is <= 0, a default of 10 is used.
// If config is nil, DefaultConfig() will be used.
func NewTCPPool(address string, config *Config, capacity int) (*EIPTCPPool, error) {
	if capacity <= 0 {
		capacity = 10
	}

	if config == nil {
		config = DefaultConfig()
	}

	return &EIPTCPPool{
		factory: func() (*EIPTCP, error) {
			conn, err := NewTCP(address, config)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		pool: make(chan *EIPTCP, capacity),
	}, nil
}

// Get retrieves a connection from the pool.
// If an idle connection is available, it is returned.
// Otherwise, a new connection is created and connected.
// The caller should return the connection to the pool using Put when done.
func (p *EIPTCPPool) Get() (*EIPTCP, error) {
	if p == nil {
		return nil, errors.New("pool is nil")
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("pool is closed")
	}
	p.mu.Unlock()

	select {
	case conn := <-p.pool:
		if conn == nil {
			break
		}
		return conn, nil
	default:
	}

	conn, err := p.factory()
	if err != nil {
		return nil, err
	}
	if err := conn.Connect(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// Put returns a connection to the pool.
// If the pool is closed, the connection is closed instead.
// If the pool is at capacity, the connection is closed.
func (p *EIPTCPPool) Put(conn *EIPTCP) error {
	if p == nil || conn == nil {
		return errors.New("pool or connection is nil")
	}

	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()

	if closed {
		return conn.Close()
	}

	select {
	case p.pool <- conn:
		return nil
	default:
		return conn.Close()
	}
}

// Close closes the pool and all connections in it.
// After Close is called, Get and Put will return errors.
// Close is idempotent and safe to call multiple times.
func (p *EIPTCPPool) Close() error {
	if p == nil {
		return errors.New("pool is nil")
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.pool)
	p.mu.Unlock()

	var firstErr error
	for conn := range p.pool {
		if conn == nil {
			continue
		}
		if err := conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
