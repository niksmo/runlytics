package counter

import "sync"

type counter struct {
	mu       sync.Mutex
	start, n int
}

// Return increment func with started at `n`
func New(start int) *counter {
	return &counter{start: start, n: start}
}

func (c *counter) Next() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
	return c.n
}

func (c *counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n = c.start
}
