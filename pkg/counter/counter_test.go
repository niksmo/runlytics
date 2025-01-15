package counter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounterNext(t *testing.T) {
	type test struct {
		name string
		want int64
	}

	tests := []test{
		{
			name: "Init is zero number",
			want: 13,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			counter := New()
			var current int64
			for range test.want {
				current = counter.Next()
			}
			assert.Equal(t, test.want, current)
		})
	}
}
