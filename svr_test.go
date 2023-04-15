package pri_dns

import (
	cidrMerger "github.com/laeni/pri-dns/cidr-merger"
	"net"
	"reflect"
	"testing"
)

func Test_isIpBefore(t *testing.T) {
	type args struct {
		a net.IP
		b net.IP
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "a < b",
			args: args{a: net.IPv4(1, 2, 1, 2), b: net.IPv4(1, 2, 1, 3)},
			want: true,
		},
		{
			name: "a = b",
			args: args{a: net.IPv4(1, 2, 1, 2), b: net.IPv4(1, 2, 1, 2)},
			want: false,
		},
		{
			name: "a > b",
			args: args{a: net.IPv4(1, 2, 1, 3), b: net.IPv4(1, 2, 1, 2)},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIpBefore(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("isIpBefore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ipPlusOne(t *testing.T) {
	tests := []struct {
		name string
		args net.IP
		want net.IP
	}{
		{
			name: "简单IP +1 - 1.2.1.2",
			args: net.IPv4(1, 2, 1, 2),
			want: net.IPv4(1, 2, 1, 3),
		},
		{
			name: "进位 +1 - 1.2.1.255",
			args: net.IPv4(1, 2, 1, 255),
			want: net.IPv4(1, 2, 2, 0),
		},
		{
			name: "进位 +1 - 1.2.255.255",
			args: net.IPv4(1, 2, 255, 255),
			want: net.IPv4(1, 3, 0, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ipPlusOne(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ipPlusOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ipMinusOne(t *testing.T) {
	tests := []struct {
		name string
		args net.IP
		want net.IP
	}{
		{
			name: "简单IP -1 - 1.2.1.1",
			args: net.IPv4(1, 2, 1, 2),
			want: net.IPv4(1, 2, 1, 1),
		},
		{
			name: "借位 -1 - 1.2.1.0",
			args: net.IPv4(1, 2, 1, 0),
			want: net.IPv4(1, 2, 0, 255),
		},
		{
			name: "借位 -1 - 1.2.0.0",
			args: net.IPv4(1, 2, 0, 0),
			want: net.IPv4(1, 1, 255, 255),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ipMinusOne(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ipMinusOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_excludeIpRange(t *testing.T) {
	type args struct {
		irs []cidrMerger.IRange
		exs []cidrMerger.IRange
	}

	tests := []struct {
		name string
		args args
		want []cidrMerger.IRange
	}{
		{
			name: "从网段中排除单个IP",
			args: args{
				irs: []cidrMerger.IRange{&cidrMerger.IpNetWrapper{IPNet: &net.IPNet{
					IP:   net.IPv4(1, 2, 3, 0),
					Mask: net.CIDRMask(24, 32),
				}}},
				exs: []cidrMerger.IRange{&cidrMerger.IpWrapper{IP: net.IPv4(1, 2, 3, 4)}},
			},
			want: []cidrMerger.IRange{
				&cidrMerger.Range{
					Start: net.IPv4(1, 2, 3, 0),
					End:   net.IPv4(1, 2, 3, 3),
				},
				&cidrMerger.Range{
					Start: net.IPv4(1, 2, 3, 5),
					End:   net.IPv4(1, 2, 3, 255),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := excludeIpRange(tt.args.irs, tt.args.exs); !iRangeEqual(got, tt.want) {
				t.Errorf("excludeIpRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func iRangeEqual(a, b []cidrMerger.IRange) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		aRange := a[i].ToRange()
		bRange := b[i].ToRange()
		if !aRange.Start.Equal(bRange.Start) || !aRange.End.Equal(bRange.End) {
			return false
		}
	}
	return true
}
