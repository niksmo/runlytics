package counter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	type test struct {
		name string
		arg  int
		want int
	}

	tests := []test{
		{
			name: "Init is zero number",
			arg:  0,
			want: 3,
		},
		{
			name: "Init is negative number",
			arg:  -10,
			want: -7,
		},
		{
			name: "Init is positive number",
			arg:  10,
			want: 13,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			count := New(test.arg)
			var current int
			for range 3 {
				current = count()
			}
			assert.Equal(t, test.want, current)
		})
	}
}
