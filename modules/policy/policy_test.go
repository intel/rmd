package policy

import (
	"reflect"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	//it is neough to have it once in a package
	pflag.String("address", "", "Listen address")
	pflag.Int("tlsport", 0, "TLS listen port")
	pflag.BoolP("debug", "d", false, "Enable debug")
	pflag.String("unixsock", "", "Unix sock file path")
	pflag.Int("debugport", 0, "Debug listen port")
	pflag.String("conf-dir", "", "Directory of config file")
	pflag.String("clientauth", "challenge", "The policy the server will follow for TLS Client Authentication")

	pflag.Parse()

	viper.Set("default.policypath", "policy_test_file.yaml")
}

func Test_loadPolicy(t *testing.T) {

	var cacheParam = Param{
		"max": 2,
		"min": 2,
	}

	var pstateParam = Param{
		"ratio": 1.5,
	}

	var testModule = Module{"cache": cacheParam, "pstate": pstateParam}
	var testPolicy = Policy{"gold": testModule}
	var testArch = CPUArchitecture{"skylake": testPolicy, "broadwell": testPolicy}

	tests := []struct {
		name    string
		want    CPUArchitecture
		wantErr bool
	}{
		{"skylake success load", testArch, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadPolicy()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefaultPlatformPolicy(t *testing.T) {

	var cacheParam = Param{"max": 2, "min": 2}
	var pstateParam = Param{"ratio": 1.5}
	var testModule = Module{"cache": cacheParam, "pstate": pstateParam}
	var testPolicy = Policy{"gold": testModule}

	tests := []struct {
		name    string
		want    Policy
		wantErr bool
	}{
		{"skylake success load", testPolicy, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDefaultPlatformPolicy()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefaultPlatformPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefaultPlatformPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadPolicyInfo(t *testing.T) {

	var cacheParam = Param{"max": 2, "min": 2}
	var pstateParam = Param{"ratio": 1.5}
	var testModule = Module{"cache": cacheParam, "pstate": pstateParam}
	var testPolicy = Policy{"gold": testModule}

	tests := []struct {
		name    string
		want    Policy
		wantErr bool
	}{
		{"skylake success load", testPolicy, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadPolicyInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadPolicyInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadPolicyInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefaultPolicy(t *testing.T) {

	var cacheParam = Param{"max": 2, "min": 2}
	var pstateParam = Param{"ratio": 1.5}
	var testModule = Module{"cache": cacheParam, "pstate": pstateParam}

	type args struct {
		policyName string
	}
	tests := []struct {
		name    string
		args    args
		want    Module
		wantErr bool
	}{
		{"skylake success load", args{"gold"}, testModule, false},
		{"skylake failed load", args{"fake"}, Module{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDefaultPolicy(tt.args.policyName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefaultPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefaultPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}
