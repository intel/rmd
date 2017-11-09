// +build integration
package integration_test

import (
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo"

	"gopkg.in/gavv/httpexpect.v1"

	"github.com/intel/rmd/lib/cache"
	"github.com/intel/rmd/test/test_helpers"
)

var cacheSchemaTemplate string = `{
	"type": "object",
	"properties": {
		"rdt": {{.bool}},
		"cqm": {{.bool}},
		"cdp": {{.bool}},
		"cdp_enable": {{.bool}},
		"cat": {{.bool}},
		"cat_enable": {{.bool}},
		"caches": {
			"type": "object",
			"properties": {
				"l3": {
					"type": "object",
					"properties": {
						"number": {{.pint}},
						"cache_ids": {"type": "array", "items": {{.uint}}}
					}
				},
				"l2": {
					"type": "object",
					"properties": {
						"number": {{.pint}},
						"cache_ids": {"type": "array", "items": {{.uint}}}
					}
				}
			}
		}
	},
	"required": ["rdt", "cqm", "cdp", "cdp_enable", "cat", "cat_enable", "caches"]
}`

// Caches is capital ? fix it ?
var cacheLevelSchemaTemplate string = `{
	"type": "object",
	"properties": {
		"number": {{.pint}},
		"Caches": {
			"type": "object"
		}
	}
}`

var _ = Describe("Caches", func() {

	var (
		v1url            string
		he               *httpexpect.Expect
		cacheSchema      string
		cacheLevelSchema string
		llc              uint32
	)

	BeforeEach(func() {
		By("set url")
		v1url = testhelpers.GetHTTPV1URL()
		he = httpexpect.New(GinkgoT(), v1url)
	})

	AfterEach(func() {
	})

	Describe("Get the new system", func() {
		Context("when request 'cache' API", func() {
			BeforeEach(func() {
				By("Set cache schemata")
				cacheSchema, _ = testhelpers.FormatByKey(cacheSchemaTemplate,
					map[string]interface{}{
						"bool": testhelpers.BoolSchema,
						"pint": testhelpers.PositiveInteger,
						"uint": testhelpers.NonNegativeInteger})
			})

			It("Should be return 200", func() {

				repos := he.GET("/cache").
					WithHeader("Content-Type", "application/json").
					Expect().
					Status(http.StatusOK).
					JSON()

				repos.Schema(cacheSchema)
			})
		})

		Context("when request 'cache' API by level", func() {
			BeforeEach(func() {
				By("Set cache level schemata")
				cacheLevelSchema, _ = testhelpers.FormatByKey(cacheLevelSchemaTemplate,
					map[string]interface{}{
						"pint": testhelpers.PositiveInteger,
					})

				llc = syscache.GetLLC()
			})

			It("Should be return 200 for llc ", func() {

				repos := he.GET("/cache/llc").
					WithHeader("Content-Type", "application/json").
					Expect().
					Status(http.StatusOK).
					JSON()
				repos.Schema(cacheLevelSchema)
			})

			It("Should be return 200 for last level", func() {

				target_lev := strconv.FormatUint(uint64(llc), 10)
				cacheUrl := "/cache/l" + target_lev

				repos := he.GET(cacheUrl).
					WithHeader("Content-Type", "application/json").
					Expect().
					Status(http.StatusOK).
					JSON()
				repos.Schema(cacheLevelSchema)
			})

			It("Should be return 400 for not last level", func() {

				badllc := llc - 1
				target_lev := strconv.FormatUint(uint64(badllc), 10)
				cacheUrl := "/cache/l" + target_lev

				he.GET(cacheUrl).
					WithHeader("Content-Type", "application/json").
					Expect().
					Status(http.StatusBadRequest)
			})

		})
	})
})
