// +build integration

package integration_test

import (
	"fmt"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testhelpers "github.com/intel/rmd/test/test_helpers"
	util "github.com/intel/rmd/utils/bitmap"
	"github.com/intel/rmd/utils/proc"
	"github.com/intel/rmd/utils/resctrl"
	"gopkg.in/gavv/httpexpect.v1"
)

var _ = Describe("Workload", func() {

	var (
		he              *httpexpect.Expect
		isMbaSupported  bool
		defaultMbaValue int
	)

	BeforeEach(func() {
		workloadUrl := v1url + "workloads"
		he = httpexpect.New(GinkgoT(), workloadUrl)
		isMbaSupported, _ = proc.IsMbaAvailable()
		if isMbaSupported {
			defaultMbaValue = 100
		} else {
			defaultMbaValue = -1
		}
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
					Pids, []string{}, 1, 1, defaultMbaValue, "")
				verifyWrokload(he, data, isMbaSupported)
			})
		})
		// MBA not supported for shared group
		Context("When request a new workload API with max_cache = min_cache and cpus", func() {
			It("Should return 500", func() {

				data := testhelpers.AssembleRequest(
					[]*os.Process{}, []string{"4-5"}, 1, 1, defaultMbaValue, "")
				verifyWrokload(he, data, isMbaSupported)
			})
		})

		Context("When request a new workload API with max_cache > min_cache and cpus", func() {
			It("Should return 200", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{}, []string{"4-5"}, 2, 1, defaultMbaValue, "")
				verifyWrokload(he, data, isMbaSupported)
			})
		})
		// MBA not supported for shared group at current version
		// TODO PQOS: change this test when shared group will be supported
		Context("When request a new workload API with max_cache = min_cache = 0 and cpus", func() {
			It("Should return 500", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{}, []string{"4-5"}, 0, 0, defaultMbaValue, "")
				he.POST("/").WithHeader("Content-Type", "application/json").
					WithJSON(data).
					Expect().
					Status(http.StatusInternalServerError)
			})
		})

		Context("When request a new workload API with pid which doesn't exist ", func() {
			It("Should return 400", func() {
				data := testhelpers.AssembleRequest(
					[]*os.Process{&os.Process{Pid: 199999}}, []string{"4-5"}, 0, 0, defaultMbaValue, "")
				he.POST("/").WithHeader("Content-Type", "application/json").
					WithJSON(data).
					Expect().
					Status(http.StatusBadRequest)
			})
		})

		Context("When request a new workload API with cache and MBA values", func() {
			It("Should return 200", func() {
				if isMbaSupported {
					data := testhelpers.AssembleRequest(
						[]*os.Process{}, []string{"4-5"}, 2, 2, 50, "")
					verifyWrokload(he, data, isMbaSupported)
				} else {
					fmt.Println("Machine does not suport MBA. So Aborting!!!!")
				}
			})
		})
	})
})

func verifyWrokload(he *httpexpect.Expect, data map[string]interface{}, isMbaSupported bool) {
	repobj := he.POST("/").WithHeader("Content-Type", "application/json").
		WithJSON(data).
		Expect().
		Status(http.StatusCreated).JSON().Object()

	fmt.Println("Response :", repobj)
	fmt.Println("CACHE: ", repobj.Value("rdt").Object().Value("cache"), data["rdt"].(map[string]interface{})["cache"])
	// print plugins only if used int data (so expected also in json result)
	if plugs, ok := data["plugins"]; ok {
		fmt.Println("PLUGINS: ", repobj.Value("plugins"), plugs)
	}

	workloadId := repobj.Value("id").String().Raw()
	cosName := repobj.Value("cos_name").String().Raw()
	resall := resctrl.GetResAssociation(nil)

	fmt.Println("Resall : ", resall, resall[cosName])

	repobj.Value("status").Equal("Successful")
	if p, ok := data["policy"]; ok {
		repobj.Value("policy").Equal(p)
	} else {
		// validate plugins only if used in data (so expected also in json result)
		if plugs, ok := data["plugins"]; ok {
			repobj.Value("plugins").Equal(plugs)
		}
		repobj.Value("rdt").Object().Value("cache").Equal(data["rdt"].(map[string]interface{})["cache"])
		if isMbaSupported {
			repobj.Value("rdt").Object().Value("mba").Equal(data["rdt"].(map[string]interface{})["mba"])
		}
	}

	res, ok := resall[cosName]
	if !ok {
		Fail(fmt.Sprintf("Resource group %s was not created correctly", cosName))
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

	if maxCache, ok := data["rdt"].(map[string]interface{})["cache"].(map[string]int)["max"]; ok && maxCache == 0 {
		repobj.Value("cos_name").Equal("shared")
	}

	// Cleanup
	he.DELETE("/"+workloadId).WithHeader("Content-Type", "application/json").
		Expect().
		Status(http.StatusOK)
}
