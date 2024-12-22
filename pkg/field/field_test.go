package field

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	type args struct {
		obj   any
		field string
	}

	type test struct {
		name    string
		args    args
		want    any
		wantErr error
	}

	vInt := 1
	pInt := &vInt

	tests := []test{
		{
			name: "Get non zero public field in struct",
			args: args{
				obj:   struct{ A, b string }{A: "foo", b: "bar"},
				field: "A",
			},
			want:    "foo",
			wantErr: nil,
		},
		{
			name: "Get non zero public field in struct pointer",
			args: args{
				obj:   &struct{ A, b string }{A: "foo", b: "bar"},
				field: "A",
			},
			want:    "foo",
			wantErr: nil,
		},
		{
			name: "Get zero public field in struct",
			args: args{
				obj:   struct{ A, b int }{A: 0, b: 1},
				field: "A",
			},
			want:    0,
			wantErr: nil,
		},
		{
			name: "Get nil public field in struct",
			args: args{
				obj:   struct{ A map[string]int }{},
				field: "A",
			},
			want:    map[string]int(nil),
			wantErr: nil,
		},
		{
			name: "Get private field in struct",
			args: args{
				obj:   struct{ A, b string }{A: "foo", b: "bar"},
				field: "b",
			},
			want:    nil,
			wantErr: ErrUnexportedField,
		},
		{
			name: "Get not exists field in struct",
			args: args{
				obj:   struct{ A, b string }{A: "foo", b: "bar"},
				field: "c",
			},
			want:    nil,
			wantErr: ErrNoField,
		},
		{
			name: "Not struct value",
			args: args{
				obj:   vInt,
				field: "A",
			},
			want:    nil,
			wantErr: ErrNotStruct,
		},
		{
			name: "Not struct pointer",
			args: args{
				obj:   pInt,
				field: "A",
			},
			want:    nil,
			wantErr: ErrNotStruct,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := Value(test.args.obj, test.args.field)

			if test.wantErr != nil {
				assert.True(t, errors.Is(err, test.wantErr))
				return
			}
			assert.Equal(t, test.want, v)
		})
	}

}
