package counter

import (
	"sync/atomic"
)

type Counter struct {
	n atomic.Int64
}

func New() *Counter {
	return &Counter{}
}

func (c *Counter) Next() int64 {
	return c.n.Add(1)
}
