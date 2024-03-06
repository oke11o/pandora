package clientpool

import (
	"sync/atomic"
)

func New(size int) *Pool {
	return &Pool{
		pool: make([]any, 0, size),
	}
}

type Pool struct {
	pool []any
	i    atomic.Uint64
}

func (p *Pool) Add(conn any) {
	p.pool = append(p.pool, conn)
}

func (p *Pool) Next() any {
	if len(p.pool) == 0 {
		return nil
	}
	i := p.i.Add(1)
	return p.pool[int(i)%len(p.pool)]
}
