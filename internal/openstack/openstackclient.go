// +build openstack

package openstack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	workload "github.com/intel/rmd/modules/workload/types"
	log "github.com/sirupsen/logrus"
)

//mutex
var lock sync.Mutex

//stores token needed for authentication
var token = ""

//request in string form needed for obtaining the token ("project" scope to handle both Swift and Glance)
var jsonStr = []byte(`{"auth":{"methods":["password"],"tennant":"admin","identity":{"methods":["password"],"password":{"user":{"name":"%s","domain": {"id":"default"},"password":"%s"}}},"scope": {"project": {"domain":{"id": "default"},"name": "admin"}}}}`)

//stores info about endpoints - "name" as key, "url" as value
var endpointsMap = make(map[string]string)

//temporary place to store "name" and "url" for endpoint
type endpointInfo struct {
	name string
	url  string
}

//getter for token
func getToken() string {
	lock.Lock()
	defer lock.Unlock()
	if token == "" {
		obtainToken()
	}
	return token
}

//obtain new token
func obtainToken() {
	filledJSONStr := fmt.Sprintf(string(jsonStr), oscfg.KeystoneLogin, oscfg.KeystonePassword)
	response, err := http.Post(oscfg.KeystoneURL, "application/json", bytes.NewBuffer([]byte(filledJSONStr)))
	if err == nil {
		token = response.Header.Get("x-subject-token")
		obtainEndpoints(response)
	} else {
		log.Error("The HTTP request failed: ", err.Error())
	}
}

// get list of existing endpoints from token response and place them in endpointsMap
func obtainEndpoints(tokenResponse *http.Response) {
	defer tokenResponse.Body.Close()
	data, err := ioutil.ReadAll(tokenResponse.Body)

	if err != nil {
		log.Error(err)
		return
	}

	result := make(map[string]map[string]interface{})
	json.Unmarshal(data, &result)

	if err != nil {
		log.Error(err)
		return
	}
	//get content for "catalog" tag which contains information about cloud services like swift and glance
	if id, ok := result["token"]["catalog"]; ok {
		singleResult := id.([]interface{})
		//iterate on "catalog" tag's content
		for _, value := range singleResult {
			//we want to get "name" and "url" from endpoint and place in endpointsMap
			//"name" will be used as a key in endpointsMap
			element := endpointInfo{}
			catalogMap := value.(map[string]interface{})

			//get name for catalog item
			if name, ok := catalogMap["name"]; ok {
				element.name = name.(string)
			}

			//get catalog items content
			if endpoints, ok := catalogMap["endpoints"]; ok {
				//iteration on a complex piece of data to get endpoint url
				for _, endpointContainer := range endpoints.([]interface{}) {

					for endpointField, endpointValue := range endpointContainer.(map[string]interface{}) {

						if endpointField == "url" {
							tempValue := strings.TrimSpace(endpointValue.(string))

							//TODO: RMD requires Glance (optional) and Swift (preferred as storage place) only,
							//during further tests decide if overwritting value in a map is a problem here.
							//Currently Openstack and RMD are run on the same computer and below rule for Swift
							//seems to be sufficient and Glance has one endpoint or its multiple endpoints
							//contains the same url value

							//for Swift endpoint we need only url which contains "AUTH" substring
							//correct url example: http://localhost:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef
							if element.name == "swift" {
								if strings.Contains(tempValue, "AUTH") == true {
									element.url = tempValue
								} else {
									continue
								}
							}
							//handling for other endpoints than Swift
							element.url = tempValue
						}
					}
				}
			}
			// add endpoint to map only when name and url are specified
			if element.name != "" && element.url != "" {
				log.Infof("Add endpoint to map - name: %s  url: %s", element.name, element.url)
				endpointsMap[element.name] = element.url
			}
		}
	} else {
		log.Error("Cannot obtain endpoints from HTTP response")
	}
}

// function to handle case when token exists but is no longer valid - HTTP 401 error code
// so we need obtain new token and resend query
func sendQuery(method string, url string) ([]byte, error) {
	data, statusCode, err := sendSingleQuery(method, url)

	if err != nil {
		log.Infof("The HTTP GET request failed with error: %s\n", err.Error())
	} else {
		if statusCode == 401 {
			obtainToken()
			data, _, err = sendSingleQuery(method, url)
		}
	}

	return data, err
}

//send query and get response
func sendSingleQuery(method string, url string) ([]byte, int, error) {
	var data []byte
	statusCode := 400 //Bad Request as default state
	request, err := http.NewRequest(method, url, nil)

	if err == nil {
		request.Header.Set("X-Auth-Token", getToken())
		client := &http.Client{Timeout: time.Second * 5}
		response, err := client.Do(request)

		if err == nil {
			defer response.Body.Close()
			data, _ = ioutil.ReadAll(response.Body)
			statusCode = response.StatusCode
		}
	} else {
		log.Errorf("Request creation failed - %s", err.Error())
	}
	return data, statusCode, err
}

// unmarshal workload policy data obtained from Glance
func unMarshallGlanceWorkloadPolicyUUID(responseData []byte) string {

	//check if empty
	if len(responseData) == 0 {
		log.Error("Cannot unmarshal empty response")
		return ""
	}

	type ExtraSpecs struct {
		GlanceWorkloadPolicyUUID string `json:"rmd:glance_workload_policy_uuid"`
	}

	type Response struct {
		ExtraSpecsField ExtraSpecs `json:"extra_specs"`
	}

	var myStruct Response
	err := json.Unmarshal(responseData, &myStruct)

	if err != nil {
		log.Error(err)
	}

	return myStruct.ExtraSpecsField.GlanceWorkloadPolicyUUID
}

func getGlanceWorkloadPolicyByUUID(imageID string) (*workload.RDTWorkLoad, error) {
	templateURLStr := "{http://localhost/v2/images}/{imageID}/file"
	filledURLStr := strings.Replace(templateURLStr, "{imageID}", imageID, 1)
	workload := new(workload.RDTWorkLoad)
	var err error
	if value, ok := endpointsMap["glance"]; ok {

		filledURLStr = strings.Replace(filledURLStr, "{http://localhost/v2/images}", value, 1)

		data, err := sendQuery("GET", filledURLStr)

		if err != nil {
			log.Error(err)
			return nil, err
		}

		err = json.Unmarshal(data, &workload)
		if err != nil {
			log.Error(err)
			return nil, err
		}

	} else {
		log.Errorf("Glance record not found in endpoint map")
		return nil, errors.New("Glance record not found in endpoint map")
	}

	return workload, err
}

// unmarshal workload policy data obtained from Swift
func unMarshallSwiftWorkloadPolicyUUID(responseData []byte) string {

	//check if empty
	if len(responseData) == 0 {
		log.Error("Cannot unmarshal empty response")
		return ""
	}

	type ExtraSpecs struct {
		SwiftWorkloadPolicyUUID string `json:"rmd:swift_workload_policy_url"`
	}

	type Response struct {
		ExtraSpecsField ExtraSpecs `json:"extra_specs"`
	}

	var myStruct Response
	err := json.Unmarshal(responseData, &myStruct)

	if err != nil {
		log.Error(err)
	}

	return myStruct.ExtraSpecsField.SwiftWorkloadPolicyUUID
}

//get workload policy UUID - support for Swift and Glance
func getWorkloadPolicyUUID(flavorID string, useSwiftInsteadOfGlance bool) string {
	result := ""
	templateURLStr := "{http://localhost/compute/v2.1}/flavors/{flavorId}/os-extra_specs"
	filledURLStr := strings.Replace(templateURLStr, "{flavorId}", flavorID, 1)

	if value, ok := endpointsMap["nova"]; ok {
		filledURLStr = strings.Replace(filledURLStr, "{http://localhost/compute/v2.1}", value, 1)

		data, err := sendQuery("GET", filledURLStr)

		if err != nil {
			log.Infof("The HTTP GET request failed with error: %s\n", err.Error())
		} else {
			if useSwiftInsteadOfGlance == true {
				log.Infof("Taking policy from Swift...")
				result = unMarshallSwiftWorkloadPolicyUUID(data)
			} else {
				log.Infof("Taking policy from Glance...")
				result = unMarshallGlanceWorkloadPolicyUUID(data)
			}
		}
	} else {
		log.Errorf("Nova record not found in endpoint map")
	}

	log.Infof("getWorkloadPolicyUUID: %s", string(result))
	return result
}

func getSwiftWorkloadPolicyByURL(workloadPolicyURL string) (*workload.RDTWorkLoad, error) {
	// full URL example = http://localhost:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef/RmdPolicies/custom.json
	// where "RmdPolicies/custom.json" is {workloadPolicyURL}
	templateURLStr := "{http://localhost:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef}/{workloadPolicyURL}"
	workload := new(workload.RDTWorkLoad)
	var err error

	if value, ok := endpointsMap["swift"]; ok {
		templateURLStr = strings.Replace(templateURLStr, "{http://localhost:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef}", value, 1)
		templateURLStr = strings.Replace(templateURLStr, "{workloadPolicyURL}", workloadPolicyURL, 1)
		log.Infof("Swift workload policy link: %s", templateURLStr)

		data, err := sendQuery("GET", templateURLStr)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		err = json.Unmarshal(data, &workload)
		if err != nil {
			log.Error(err)
			return nil, err
		}

	} else {
		log.Error("Swift record not found in endpoint map")
		return nil, errors.New("Swift record not found in endpoint map")
	}

	return workload, err
}
