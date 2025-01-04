package counter

// Return increment func with started at `n`
func New(n int) func() int {
	return func() int {
		n++
		return n
	}
}
