package util

import (
	"reflect"
	"testing"
)

func TestSliceRemoveItem(t *testing.T) {
	type args[T interface{}] struct {
		s []T
		i int
	}
	type testCase[T interface{}] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[string]{
		{
			name: "",
			args: args[string]{[]string{"a", "b", "c", "d"}, 2},
			want: []string{"a", "b", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliceRemoveItem(tt.args.s, tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SliceRemoveItem() = %v, want %v", got, tt.want)
			}
		})
	}
}
