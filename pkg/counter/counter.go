// Package counter provides simple concurrency safety counter.
package counter

import (
	"sync/atomic"
)

// Counter abstraction over [atomic.Int64]
type Counter struct {
	n atomic.Int64
}

// New returns Counter pointer.
func New() *Counter {
	return &Counter{}
}

// Next incrementing underlying number and returns new value.
func (c *Counter) Next() int64 {
	return c.n.Add(1)
}
