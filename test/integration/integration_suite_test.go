// +build integration
package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
	"testing"

	"github.com/intel/rmd/test/test_helpers"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var v1url string
var PidNumber = 5
var Pids []*os.Process

var _ = BeforeSuite(func() {
	err := testhelpers.ConfigInit(os.Getenv("CONF"))
	Expect(err).NotTo(HaveOccurred())

	Pids, err = testhelpers.CreateNewProcesses("sleep 100", PidNumber)
	Expect(err).NotTo(HaveOccurred())

	v1url = testhelpers.GetHTTPV1URL()
})

var _ = AfterSuite(func() {
	//Cleanup processes
	testhelpers.CleanupProcesses(Pids)
})
