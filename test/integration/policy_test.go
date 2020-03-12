// +build integration

package integration_test

import (
	"github.com/intel/rmd/modules/policy"
	testhelpers "github.com/intel/rmd/test/test_helpers"
	. "github.com/onsi/ginkgo"
	"gopkg.in/gavv/httpexpect.v1"
	"net/http"
)

var _ = Describe("Policy", func() {

	var (
		v1url    string
		he       *httpexpect.Expect
		policies policy.Policy
	)

	BeforeEach(func() {
		By("set url")
		v1url = testhelpers.GetHTTPV1URL()
		he = httpexpect.New(GinkgoT(), v1url)
	})

	AfterEach(func() {
	})

	Describe("Get the new system", func() {
		Context("when request 'policy' API", func() {
			BeforeEach(func() {
			})

			It("Should be return 200", func() {
				// policy returns an array
				policies, _ = policy.LoadPolicyInfo()
				reparr := he.GET("/policy").
					WithHeader("Content-Type", "application/json").
					Expect().
					Status(http.StatusOK).
					JSON()
				reparr.NotNull()
				reparr.Equal(policies)
			})
		})
	})
})
