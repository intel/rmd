// +build integration
package integration_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"

	"github.com/spf13/viper"
	"gopkg.in/gavv/httpexpect.v1"

	"github.com/intel/rmd/lib/cpu"
	"github.com/intel/rmd/model/policy"
	"github.com/intel/rmd/test/test_helpers"
)

var _ = Describe("Policy", func() {

	var (
		v1url    string
		he       *httpexpect.Expect
		policies []policy.Policy
	)

	BeforeEach(func() {
		By("set url")
		v1url = testhelpers.GetHTTPV1URL()
		he = httpexpect.New(GinkgoT(), v1url)
		policyPath := testhelpers.GetPolicyPath()

		configFileExt := filepath.Ext(policyPath)
		configType := strings.TrimPrefix(configFileExt, ".")
		r, _ := ioutil.ReadFile(policyPath)

		runtime_viper := viper.New()
		runtime_viper.SetConfigType(configType)
		runtime_viper.ReadConfig(bytes.NewBuffer(r)) // Find and read the config file
		c := policy.CATConfig{}
		runtime_viper.Unmarshal(&c)
		platform := cpu.GetMicroArch(cpu.GetSignature())

		// Grab polices from config file
		policies = c.Catpolicy[strings.ToLower(platform)]
	})

	AfterEach(func() {
	})

	Describe("Get the new system", func() {
		Context("when request 'policy' API", func() {
			BeforeEach(func() {
			})

			It("Should be return 200", func() {

				// policy returns an array
				reparr := he.GET("/policy").
					WithHeader("Content-Type", "application/json").
					Expect().
					Status(http.StatusOK).
					JSON().Array()

				reparr.NotEmpty()
				reparr.Equal(policies)
			})
		})
	})
})
