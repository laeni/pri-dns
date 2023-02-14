package util

import (
	"reflect"
	"testing"
)

func TestReverseDomain(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "测试 example.com",
			args: args{str: "example.com"},
			want: "com.example",
		},
		{
			name: "测试 a.example.cn",
			args: args{str: "a.example.cn"},
			want: "cn.example.a",
		},
		{
			name: "测试 example.com.cn",
			args: args{str: "example.com.cn"},
			want: "com.cn.example",
		},
		{
			name: "测试 example.com.",
			args: args{str: "example.com."},
			want: "com.example",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReverseDomain(tt.args.str); got != tt.want {
				t.Errorf("ReverseDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			want: []string{"com.example.b.a", "com.example.b.a.*", "com.example.b.*", "com.example.*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenAllMatchDomain(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenAllMatchDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
