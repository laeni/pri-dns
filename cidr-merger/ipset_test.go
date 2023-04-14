package cidr_merger

import (
	"net"
	"reflect"
	"testing"
)

func TestRange_ToIpNets(t *testing.T) {
	type fields struct {
		start net.IP
		end   net.IP
	}
	tests := []struct {
		name   string
		fields fields
		want   []*net.IPNet
	}{
		{
			name:   "0.0.0.0 - 255.0.0.0",
			fields: fields{start: net.ParseIP("0.0.0.0"), end: net.ParseIP("255.0.0.0")},
			want: []*net.IPNet{
				{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)},
				{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Range{
				Start: tt.fields.start,
				End:   tt.fields.end,
			}
			if got := r.ToIpNets(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToIpNets() = %v, want %v", got, tt.want)
			}
		})
	}
}
