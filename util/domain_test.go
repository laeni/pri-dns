package util

import (
	"reflect"
	"testing"
)

// TestReverseDomainAndToSlice tests the ReverseDomainAndToSlice function.
func TestReverseDomainAndToSlice(t *testing.T) {
	testCases := []struct {
		domain string
		want   []string
	}{
		{"example.com", []string{"com", "example"}},
		{"www.example.com", []string{"com", "example", "www"}},
		{"example.com.cn", []string{"cn", "com", "example"}},
		{"www.example.com.cn", []string{"cn", "com", "example", "www"}},
		{"", []string{}},
		{"cn", []string{"cn"}},
	}

	for _, tc := range testCases {
		t.Run(tc.domain, func(t *testing.T) {
			got := ReverseDomainAndToSlice(tc.domain)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ReverseDomainAndToSlice(%q) = %v; want %v", tc.domain, got, tc.want)
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
			// [* a.b.example.com com.example.b.a a.b.example.com.* *.com b.example.com.* *.com.example example.com.* *.com.example.b com.* *.com.example.b.a]
			want: []string{"a.b.example.com", "*.a.b.example.com", "*.b.example.com", "*.example.com", "*.com", "*",
				"com.example.b.a", "com.example.b.a.*", "com.example.b.*", "com.example.*", "com.*"},
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
