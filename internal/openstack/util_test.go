// + build linux,openstack

package openstack

import (
	"reflect"
	"strconv"
	"testing"
)

func Test_parseTasksetMask(t *testing.T) {
	cores := []string{}
	for i := 0; i < 2048; i++ {
		cores = append(cores, strconv.Itoa(i))
	}
	mask := ""
	for i := 0; i < 512; i++ {
		mask += "F"
	}

	type args struct {
		mask string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"8 bits set", args{"FF"}, []string{"0", "1", "2", "3", "4", "5", "6", "7"}},
		{"1 bits set", args{"01"}, []string{"0"}},
		{"1 bits set", args{"0E"}, []string{"1", "2", "3"}},
		{"2 bits set", args{"11"}, []string{"0", "4"}},
		{"2048 bits set", args{mask}, cores},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTasksetMask(tt.args.mask); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTasksetMask() = %v, want %v", got, tt.want)
			}
		})
	}
}
