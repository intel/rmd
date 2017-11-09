// +build integrationHTTPS

package integrationHTTPS

import (
	"crypto/tls"
	"github.com/intel/rmd/test/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/gavv/httpexpect.v1"
	"net/http"
	"os"
	"testing"
)

var (
	he *httpexpect.Expect
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration HTTPS Suite")
}

var _ = BeforeSuite(func() {

	err := testhelpers.ConfigInit(os.Getenv("CONF"))
	Expect(err).NotTo(HaveOccurred())

	skipVerify := false
	if testhelpers.GetClientAuthType() == "no" {
		skipVerify = true
	}

	he = httpexpect.WithConfig(httpexpect.Config{
		BaseURL: testhelpers.GetHTTPSV1URL(),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipVerify},
			},
		},
		Reporter: httpexpect.NewAssertReporter(
			httpexpect.NewAssertReporter(GinkgoT()),
		),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(GinkgoT(), true),
		},
	})

})
