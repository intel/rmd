package flock_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFlock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flock Suite")
}
