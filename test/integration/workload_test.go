// +build integration
package integration_test

import (
	"fmt"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"

	"gopkg.in/gavv/httpexpect.v1"
	"github.com/intel/rmd/test/test_helpers"
)

var _ = Describe("Workload", func() {

	var (
		he *httpexpect.Expect
	)

	BeforeEach(func() {
		workloadUrl := v1url + "workloads"
		he = httpexpect.New(GinkgoT(), workloadUrl)
	})

	AfterEach(func() {
	})

	Describe("Get the new system", func() {
		Context("No workload was created", func() {
			It("Should return empty response", func() {
				he.GET("/").WithHeader("Content-Type", "application/json").
					Expect().
					Status(http.StatusOK).JSON().Array().Empty()
			})
		})

		Context("When request a new workload API with max_cache = min_cache and task id", func() {
			It("Should return 200", func() {
				data := testhelpers.AssembleRequest(
					Pids, []string{}, 1, 1, "")
				verifyWrokload(he, data)
			})
		})

		Context("When request a new workload API with max_cache = min_cache and cpus", func() {
			It("Should return 200", func() {

				data := testhelpers.AssembleRequest(
					[]*os.Process{}, []string{"20-21"}, 1, 1, "")
				verifyWrokload(he, data)
			})
		})

		Context("When request a new workload API with max_cache > min_cache and cpus", func() {
			It("Should return 200", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{}, []string{"20-21"}, 2, 1, "")
				verifyWrokload(he, data)
			})
		})

		Context("When request a new workload API with max_cache = min_cache = 0 and cpus", func() {
			It("Should return 200", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{}, []string{"20-21"}, 0, 0, "")
				verifyWrokload(he, data)
			})
		})

		Context("When request a new workload API with pid which doesn't exist ", func() {
			It("Should return 400", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{&os.Process{Pid: 199999}}, []string{"20-21"}, 0, 0, "")
				he.POST("/").WithHeader("Content-Type", "application/json").
					WithJSON(data).
					Expect().
					Status(http.StatusBadRequest)
			})
		})
	})
})

func verifyWrokload(he *httpexpect.Expect, data map[string]interface{}) {

	repobj := he.POST("/").WithHeader("Content-Type", "application/json").
		WithJSON(data).
		Expect().
		Status(http.StatusCreated).JSON().Object()

	workloadId := repobj.Value("id").String().Raw()
	cosName := repobj.Value("cos_name").String().Raw()
	repobj.Value("status").Equal("Successful")

	if p, ok := data["policy"]; ok {
		repobj.Value("policy").Equal(p)
	} else {
		repobj.Value("max_cache").Equal(data["max_cache"])
		repobj.Value("min_cache").Equal(data["min_cache"])
	}

	if tids, ok := data["task_ids"]; ok {
		repobj.Value("task_ids").Equal(tids)
	} else {
		repobj.Value("core_ids").Equal(data["core_ids"])
	}

	if maxCache, ok := data["max_cache"]; ok && maxCache == 0 {
		repobj.Value("cos_name").Equal("shared")
	}

	he.DELETE("/"+workloadId).WithHeader("Content-Type", "application/json").
		Expect().
		Status(http.StatusOK)

	// TODO: verify schemata and taskids are there
	fmt.Println(cosName)

}
