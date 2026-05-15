package ethernet_ip

import (
	"errors"
	"sync"
)

type EIPTCPPool struct {
	factory func() (*EIPTCP, error)
	pool    chan *EIPTCP
	mu      sync.Mutex
	closed  bool
}

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
