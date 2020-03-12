// +build integration

package integration_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"os"

	testhelpers "github.com/intel/rmd/test/test_helpers"
	util "github.com/intel/rmd/utils/bitmap"
	"github.com/intel/rmd/utils/resctrl"
	"gopkg.in/gavv/httpexpect.v1"
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
					[]*os.Process{}, []string{"4-5"}, 1, 1, "")
				verifyWrokload(he, data)
			})
		})

		Context("When request a new workload API with max_cache > min_cache and cpus", func() {
			It("Should return 200", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{}, []string{"4-5"}, 2, 1, "")
				verifyWrokload(he, data)
			})
		})

		Context("When request a new workload API with max_cache = min_cache = 0 and cpus", func() {
			It("Should return 200", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{}, []string{"4-5"}, 0, 0, "")
				verifyWrokload(he, data)
			})
		})

		Context("When request a new workload API with pid which doesn't exist ", func() {
			It("Should return 400", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{&os.Process{Pid: 199999}}, []string{"4-5"}, 0, 0, "")
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

	fmt.Println("Response :", repobj)
	fmt.Println("CACHE: ", repobj.Value("cache"), data["cache"])

	workloadId := repobj.Value("id").String().Raw()
	cosName := repobj.Value("cos_name").String().Raw()
	resall := resctrl.GetResAssociation()

	fmt.Println("Resall : ", resall, resall[cosName])

	repobj.Value("status").Equal("Successful")
	if p, ok := data["policy"]; ok {
		repobj.Value("policy").Equal(p)
	} else {
		repobj.Value("cache").Equal(data["cache"])
	}

	res, ok := resall[cosName]
	if !ok {
		Fail(fmt.Sprintf("Resource group %s was not created correctlly", cosName))
	}

	if tids, ok := data["task_ids"]; ok {
		repobj.Value("task_ids").Equal(tids)
		Ω(res.Tasks).Should(Equal(tids))
	} else {
		fmt.Println("Core_ids : ", repobj.Value("core_ids"), data["core_ids"])
		repobj.Value("core_ids").Equal(data["core_ids"])
		fmt.Println("Core_ids : ", repobj.Value("core_ids"), data["core_ids"])
		cpubm, _ := util.NewBitmap(data["core_ids"])
		rescpubm, _ := util.NewBitmap(res.CPUs)
		fmt.Println("res.CPUs:", res.CPUs)
		fmt.Println("data", data["core_ids"])
		fmt.Println("cpubm", cpubm)
		fmt.Println("rescpubm", rescpubm)
		fmt.Println("resc", rescpubm.ToHumanString())
		fmt.Println("cpu", cpubm.ToHumanString())
		Ω(rescpubm.ToHumanString()).Should(Equal(cpubm.ToHumanString()))
	}

	if maxCache, ok := data["cache"].(map[string]int)["max"]; ok && maxCache == 0 {
		repobj.Value("cos_name").Equal("shared")
	}

	// Cleanup
	he.DELETE("/"+workloadId).WithHeader("Content-Type", "application/json").
		Expect().
		Status(http.StatusOK)
}
