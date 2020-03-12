// +build integration

package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"os"
	"testing"

	"github.com/intel/rmd/test/test_helpers"
	"github.com/intel/rmd/utils/config"
	"github.com/intel/rmd/utils/flag"
	"github.com/intel/rmd/utils/resctrl"
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
	flag.InitFlags()

	if err := config.Init(); err != nil {
		fmt.Println("Init config failed:", err)
		os.Exit(1)
	}
	if err := resctrl.Init(); err != nil {
		fmt.Println("Init config failed:", err)
		os.Exit(1)
	}

	Pids, err = testhelpers.CreateNewProcesses("sleep 100", PidNumber)
	Expect(err).NotTo(HaveOccurred())

	v1url = testhelpers.GetHTTPV1URL()
})

var _ = AfterSuite(func() {
	//Cleanup processes
	testhelpers.CleanupProcesses(Pids)
})
