package util

import (
	"testing"
)

func TestIsZeroHexString(t *testing.T) {

	cpus := "000000,00000000,00000000"
	if !IsZeroHexString(cpus) {
		t.Errorf("Misjudgement, '%s' is a string with all zero.\n", cpus)
	}

	cpus = "000000,00000000,00000001"
	if IsZeroHexString(cpus) {
		t.Errorf("Misjudgement, '%s' is not a string with all zero.\n", cpus)
	}
}
