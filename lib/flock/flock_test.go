package flock_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/intel/rmd/lib/flock"
)

var _ = Describe("Flock", func() {

	var (
		fname string
		f     *os.File
		err   error
	)

	BeforeEach(func() {
		fname = "./test.txt"
	})

	AfterEach(func() {
		Funlock(f)
		f.Close()
		os.Remove(fname)
	})

	// Separating Creation and Configuration
	JustBeforeEach(func() {
		flag := os.O_RDWR | os.O_CREATE
		f, err = os.OpenFile(fname, flag, 0600)
		if err != nil {
			Fail("Failed to open file")
		}
		err = Flock(f, 1000*1000*1000, true)
	})

	Describe("Showing how to use Flock", func() {
		It("should not error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
