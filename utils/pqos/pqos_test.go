package pqos

import (
	"os"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

// WARNING: Tests only for functions that don't use C library (no cgo-based tests)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func TestUseAvailableCLOS(t *testing.T) {

	usedCLOSes = []string{}
	availableCLOSes = []string{"COS2", "COS3", "COS4"}

	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"Positive (COS2)", "COS2", false},
		{"Positive (COS3)", "COS3", false},
		{"Positive (COS4)", "COS4", false},
		{"Negative ", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UseAvailableCLOS()
			if (err != nil) != tt.wantErr {
				t.Errorf("UseAvailableCLOS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UseAvailableCLOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNumberOfFreeCLOSes(t *testing.T) {
	tests := []struct {
		availableList []string
		name          string
		want          int
	}{
		{[]string{"COS2", "COS3"}, "2 COSes", 2},
		{[]string{}, "No COSes", 0},
		{[]string{"COS4"}, "1 COS", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availableCLOSes = tt.availableList
			if got := GetNumberOfFreeCLOSes(); got != tt.want {
				t.Errorf("GetNumberOfFreeCLOSes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReturnClos(t *testing.T) {

	usedCLOSes = []string{"COS5", "COS6", "COS7", "COS8"}
	availableCLOSes = []string{"COS2", "COS3", "COS4"}

	tests := []struct {
		name            string
		cosname         string
		expectUsed      []string
		expectAvailable []string
		wantErr         bool
	}{
		{"Positive 1 (return COS7)", "COS7", []string{"COS5", "COS6", "COS8"},
			[]string{"COS2", "COS3", "COS4", "COS7"}, false},
		{"Negative 1 (return COS123)", "COS123", []string{"COS5", "COS6", "COS8"},
			[]string{"COS2", "COS3", "COS4", "COS7"}, true},
		{"Positive 2 (return COS5)", "COS5", []string{"COS6", "COS8"},
			[]string{"COS2", "COS3", "COS4", "COS7", "COS5"}, false},
		{"Negative 2 (return COS5 again)", "COS5", []string{"COS6", "COS8"},
			[]string{"COS2", "COS3", "COS4", "COS7", "COS5"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ReturnClos(tt.cosname); (err != nil) != tt.wantErr {
				t.Errorf("ReturnClos() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.expectUsed, usedCLOSes) {
				t.Errorf("ReturnClos() error: expected used list %v found %v", tt.expectUsed, usedCLOSes)
			}
			if !reflect.DeepEqual(tt.expectAvailable, availableCLOSes) {
				t.Errorf("ReturnClos() error: expected available list %v found %v", tt.expectAvailable, availableCLOSes)
			}
		})
	}
	if len(usedCLOSes) != 2 || len(availableCLOSes) != 5 {
		t.Errorf("Invalid length of internal slices: used %v / available %v", len(usedCLOSes), len(availableCLOSes))
	}
}

func TestGetSharedCLOS(t *testing.T) {
	sharedCLOS = ""
	tests := []struct {
		name     string
		avCloses []string
		want     string
		wantErr  bool
	}{
		{"Negative", []string{}, "", true},
		{"Positive 1 (new COS reservation)", []string{"COS7"}, "COS7", false},
		{"Positive 2 (re-usage of previous COS)", []string{}, "COS7", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availableCLOSes = tt.avCloses
			got, err := GetSharedCLOS()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSharedCLOS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSharedCLOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAvailableCLOSes(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"Empty CLOS list", []string{}, []string{}},
		{"One elem", []string{"COS1"}, []string{"COS1"}},
		{"Two elems", []string{"COS1", "COS2"}, []string{"COS1", "COS2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availableCLOSes = tt.input
			if got := GetAvailableCLOSes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAvailableCLOSes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUsedCLOSes(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"Empty CLOS list", []string{}, []string{}},
		{"One elem", []string{"COS1"}, []string{"COS1"}},
		{"Two elems", []string{"COS1", "COS2"}, []string{"COS1", "COS2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usedCLOSes = tt.input
			if got := GetUsedCLOSes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUsedCLOSes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarkCLOSasUsed(t *testing.T) {

	usedCLOSes = []string{}
	tests := []struct {
		name            string
		param           string
		available       []string
		expectAvailable []string
		expectUsed      []string
		wantErr         bool
	}{
		{"Positive (1 COS available)", "COS3", []string{"COS3"},
			[]string{}, []string{"COS3"}, false},
		{"Positive (3 COSes available)", "COS5", []string{"COS4", "COS5", "COS6"},
			[]string{"COS4", "COS6"}, []string{"COS3", "COS5"}, false},
		{"Negative (no COS available)", "COS2", []string{}, []string{}, []string{"COS3", "COS5"}, true},
		{"Negative (invalid COS name)", "CCOOSS4", []string{"COS2", "COS3", "COS4", "COS5"},
			[]string{"COS2", "COS3", "COS4", "COS5"}, []string{"COS3", "COS5"}, true},
		{"Negative (COS already used)", "COS6", []string{"COS2", "COS3", "COS4", "COS5"},
			[]string{"COS2", "COS3", "COS4", "COS5"}, []string{"COS3", "COS5"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availableCLOSes = tt.available
			if err := MarkCLOSasUsed(tt.param); (err != nil) != tt.wantErr {
				t.Errorf("MarkCLOSasUsed() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.expectAvailable, availableCLOSes) {
				t.Errorf("MarkCLOSasUsed() error: expected available list %v found %v", tt.expectAvailable, availableCLOSes)
			}
			if !reflect.DeepEqual(tt.expectUsed, usedCLOSes) {
				t.Errorf("MarkCLOSasUsed() error: expected available list %v found %v", tt.expectUsed, usedCLOSes)
			}
		})
	}
}
