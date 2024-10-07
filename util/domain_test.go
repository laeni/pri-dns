package util

import (
	"testing"
)

func Test_genAllMatchDomain(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "测试 'a.b.example.com'",
			args: args{"a.b.example.com"},
			want: []string{"*", "a.b.example.com", "*.a.b.example.com", "*.b.example.com", "*.example.com", "*.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenAllMatchDomain(tt.args.name); !SliceEqual(got, tt.want) {
				t.Errorf("GenAllMatchDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
