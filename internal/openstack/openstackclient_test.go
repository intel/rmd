// +build openstack

package openstack

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	workload "github.com/intel/rmd/modules/workload/types"
	"github.com/jarcoal/httpmock"
)

func init() {
	oscfg.KeystoneURL = "http://10.237.214.102/identity/v3/auth/tokens"
	endpointsMap["nova"] = "http://localhost/compute/v2.1"
	endpointsMap["swift"] = "http://10.237.214.102:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef"
}

func Test_obtainEndpoints(t *testing.T) {

	body := `
	{
		"token":{

		"catalog":[
			{
				"endpoints":[
					{
					"region_id":"RegionOne",
					"url":"http://10.237.214.102:9696/",
					"region":"RegionOne",
					"interface":"public",
					"id":"c88aab00e06547d5b60ea0692a94e4d7"
					}
				],
				"type":"network",
				"id":"0406460b1c0447489c2068a7cbb162ee",
				"name":"neutron"
			} ]
		}
	}`

	myResponse := &http.Response{
		Status:        "201 CREATED",
		StatusCode:    201,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       nil,
		Header:        make(http.Header, 0),
	}

	//empty response
	body2 := `
	{

	}`

	myResponse2 := &http.Response{
		Status:        "201 CREATED",
		StatusCode:    201,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body2)),
		ContentLength: int64(len(body2)),
		Request:       nil,
		Header:        make(http.Header, 0),
	}

	//empty "catalog" tag
	body3 := `
	{
		"token":{

		"catalog":[
			 ]
		}
	}`

	myResponse3 := &http.Response{
		Status:        "201 CREATED",
		StatusCode:    201,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body3)),
		ContentLength: int64(len(body3)),
		Request:       nil,
		Header:        make(http.Header, 0),
	}

	//malformed "catalog" tag
	body4 := `
	{
		"token":{

		"catalog":
			 ]
		}
	}`

	myResponse4 := &http.Response{
		Status:        "201 CREATED",
		StatusCode:    201,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body4)),
		ContentLength: int64(len(body4)),
		Request:       nil,
		Header:        make(http.Header, 0),
	}

	body5 := `
	{
		"token":{

		"catalog":[
			{
				"endpoints":[
					{
					"region_id":"RegionOne",
					"url":"http://10.237.214.102:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef",
					"region":"RegionOne",
					"interface":"public",
					"id":"a5d88c02216044acb5a9324822c397a0"
					}
				],
				"type":"object-store",
				"id":"655e3ee8b8384f119c8f4e7899208888",
				"name":"swift"
			} ]
		}
	}`

	myResponse5 := &http.Response{
		Status:        "201 CREATED",
		StatusCode:    201,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body5)),
		ContentLength: int64(len(body5)),
		Request:       nil,
		Header:        make(http.Header, 0),
	}

	body6 := `
	{
		"token":{

		"catalog":[
			{
				"endpoints":[
					{
					"region_id":"RegionOne",
					"url":"http://10.237.214.102:8080/v1/0ff337bec7d4415dbc60ffca09db14ef",
					"region":"RegionOne",
					"interface":"public",
					"id":"a5d88c02216044acb5a9324822c397a0"
					}
				],
				"type":"object-store",
				"id":"655e3ee8b8384f119c8f4e7899208888",
				"name":"swift"
			} ]
		}
	}`

	myResponse6 := &http.Response{
		Status:        "201 CREATED",
		StatusCode:    201,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body6)),
		ContentLength: int64(len(body6)),
		Request:       nil,
		Header:        make(http.Header, 0),
	}

	type args struct {
		tokenResponse *http.Response
	}
	tests := []struct {
		name string
		args args
	}{
		{"Correct response without Swift", args{myResponse}},
		{"Empty response", args{myResponse2}},
		{"Empty catalog tag", args{myResponse3}},
		{"Empty catalog tag", args{myResponse4}},
		{"Correct response with Swift with AUTH", args{myResponse5}},
		{"Correct response with Swift without AUTH", args{myResponse6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obtainEndpoints(tt.args.tokenResponse)
		})
	}
}

func Test_unMarshallGlanceWorkloadPolicyUUID(t *testing.T) {

	correctCase := []byte(`{"extra_specs": {"rmd:swift_workload_policy_url": "RmdPolicies/custom.json", "rmd:glance_workload_policy_uuid": "df1a5223-e297-40d9-9a58-32387a58bed7"}}`)
	noGlanceCase := []byte(`{"extra_specs": {"rmd:swift_workload_policy_url": "RmdPolicies/custom.json"}}`)
	malformedDataCase := []byte(`{"extra_specs":`)
	emptyDataCase := []byte("")

	type args struct {
		responseData []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{

		{"Correct case", args{correctCase}, "df1a5223-e297-40d9-9a58-32387a58bed7"},
		{"No Glance case", args{noGlanceCase}, ""},
		{"Malformed data case", args{malformedDataCase}, ""},
		{"Empty data case", args{emptyDataCase}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unMarshallGlanceWorkloadPolicyUUID(tt.args.responseData); got != tt.want {
				t.Errorf("unMarshallGlanceWorkloadPolicyUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sendSingleQuery(t *testing.T) {

	correctCase := []byte(`{"extra_specs": {"hw:cpu_policy": "shared", "rmd:glance_workload_policy_uuid": "123456678", "hw:cpu_thread_policy": "prefer"}}`)
	correctURL := "http://localhost/compute/v2.1/flavors/c1/os-extra_specs"

	body := `
	{
		"token":{

		"catalog":[
			{
				"endpoints":[
					{
					"region_id":"RegionOne",
					"url":"http://10.237.214.102:9696/",
					"region":"RegionOne",
					"interface":"public",
					"id":"c88aab00e06547d5b60ea0692a94e4d7"
					}
				],
				"type":"network",
				"id":"0406460b1c0447489c2068a7cbb162ee",
				"name":"neutron"
			} ]
		}
	}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost/identity/v3/auth/tokens",
		httpmock.NewStringResponder(200, body))

	httpmock.RegisterResponder("GET", "http://localhost/compute/v2.1/flavors/c1/os-extra_specs",
		httpmock.NewStringResponder(200, `{"extra_specs": {"hw:cpu_policy": "shared", "rmd:glance_workload_policy_uuid": "123456678", "hw:cpu_thread_policy": "prefer"}}`))

	type args struct {
		method string
		url    string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		want1   int
		wantErr bool
	}{
		{"Correct case", args{"GET", correctURL}, correctCase, 200, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := sendSingleQuery(tt.args.method, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("sendSingleQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sendSingleQuery() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("sendSingleQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_sendQuery(t *testing.T) {

	correctCase := []byte(`{"extra_specs": {"hw:cpu_policy": "shared", "rmd:glance_workload_policy_uuid": "123456678", "hw:cpu_thread_policy": "prefer"}}`)
	correctURL := "http://localhost/compute/v2.1/flavors/c1/os-extra_specs"
	wrongURL := "http://localhost/compute/v2.1/flavors/abcd/os-extra_specs"

	body := `
	{
		"token":{

		"catalog":[
			{
				"endpoints":[
					{
					"region_id":"RegionOne",
					"url":"http://10.237.214.102:9696/",
					"region":"RegionOne",
					"interface":"public",
					"id":"c88aab00e06547d5b60ea0692a94e4d7"
					}
				],
				"type":"network",
				"id":"0406460b1c0447489c2068a7cbb162ee",
				"name":"neutron"
			} ]
		}
	}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://localhost/compute/v2.1/flavors/c1/os-extra_specs",
		httpmock.NewStringResponder(200, `{"extra_specs": {"hw:cpu_policy": "shared", "rmd:glance_workload_policy_uuid": "123456678", "hw:cpu_thread_policy": "prefer"}}`))

	httpmock.RegisterResponder("GET", "http://localhost/compute/v2.1/flavors/abcd/os-extra_specs",
		httpmock.NewStringResponder(401, ``))

	httpmock.RegisterResponder("POST", "http://localhost/identity/v3/auth/tokens",
		httpmock.NewStringResponder(200, body))

	type args struct {
		method string
		url    string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Correct case", args{"GET", correctURL}, correctCase, false},
		{"401 case", args{"GET", wrongURL}, []byte(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sendQuery(tt.args.method, tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("sendQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sendQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_obtainToken(t *testing.T) {

	body := `
	{
		"token":{

		"catalog":[
			{
				"endpoints":[
					{
					"region_id":"RegionOne",
					"url":"http://10.237.214.102:9696/",
					"region":"RegionOne",
					"interface":"public",
					"id":"c88aab00e06547d5b60ea0692a94e4d7"
					}
				],
				"type":"network",
				"id":"0406460b1c0447489c2068a7cbb162ee",
				"name":"neutron"
			} ]
		}
	}`
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost/identity/v3/auth/tokens",
		httpmock.NewStringResponder(200, body))

	tests := []struct {
		name string
	}{
		{"MyExampleToken"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obtainToken()
		})
	}
}

func Test_getGlanceWorkloadPolicyByUUID(t *testing.T) {

	body := `
	{
		"token":{

		"catalog":[
			 ]
		}
	}`

	testWorkload := &workload.RDTWorkLoad{
		Policy: "gold",
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost/identity/v3/auth/tokens",
		httpmock.NewStringResponder(200, body))

	httpmock.RegisterResponder("GET", "http://localhost/image/v2/images/fakeID/file",
		httpmock.NewStringResponder(404, "{}"))

	httpmock.RegisterResponder("GET", "http://localhost/image/realID/file",
		httpmock.NewStringResponder(200, `{"policy":"gold"}`))

	httpmock.RegisterResponder("GET", "http://localhost/image/wrong/file",
		httpmock.NewStringResponder(200, `{"polkscy":}`))

	type args struct {
		imageID string
	}
	tests := []struct {
		name                string
		args                args
		wantWorkload        *workload.RDTWorkLoad
		wantErr             bool
		endpointExistsInMap bool
	}{

		{"Error case", args{"fakeID"}, nil, true, false},
		{"Correct case", args{"realID"}, testWorkload, false, true},
		{"Correct case", args{"wrong"}, nil, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.endpointExistsInMap == true {
				//for some test we want to have proper record in endpointsMap
				endpointsMap["glance"] = "http://localhost/image"
			}
			gotWorkload, err := getGlanceWorkloadPolicyByUUID(tt.args.imageID)

			if (err != nil) != tt.wantErr {
				t.Errorf("getGlanceWorkloadPolicyByUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotWorkload, tt.wantWorkload) {
				t.Errorf("getGlanceWorkloadPolicyByUUID() = %v, want %v", gotWorkload, tt.wantWorkload)
			}

			if tt.endpointExistsInMap == true {
				//remove not needed record from endpointsMap
				_, ok := endpointsMap["glance"]
				if ok {
					delete(endpointsMap, "glance")
				}

			}
		})
	}
}

func Test_unMarshallSwiftWorkloadPolicyUUID(t *testing.T) {

	correctCase := []byte(`{"extra_specs": {"rmd:swift_workload_policy_url": "RmdPolicies/custom.json", "rmd:glance_workload_policy_uuid": "df1a5223-e297-40d9-9a58-32387a58bed7"}}`)
	noSwiftCase := []byte(`{"extra_specs": {"rmd:glance_workload_policy_uuid": "df1a5223-e297-40d9-9a58-32387a58bed7"}}`)
	malformedDataCase := []byte(`{"extra_specs":`)
	emptyDataCase := []byte("")

	type args struct {
		responseData []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Correct case", args{correctCase}, "RmdPolicies/custom.json"},
		{"No Swift case", args{noSwiftCase}, ""},
		{"Malformed data case", args{malformedDataCase}, ""},
		{"Empty data case", args{emptyDataCase}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unMarshallSwiftWorkloadPolicyUUID(tt.args.responseData); got != tt.want {
				t.Errorf("unMarshallSwiftWorkloadPolicyUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getWorkloadPolicyUUID(t *testing.T) {

	body := `
	{
		"token":{

		"catalog":[
			{
				"endpoints":[
					{
					"region_id":"RegionOne",
					"url":"http://10.237.214.102:9696/",
					"region":"RegionOne",
					"interface":"public",
					"id":"c88aab00e06547d5b60ea0692a94e4d7"
					}
				],
				"type":"network",
				"id":"0406460b1c0447489c2068a7cbb162ee",
				"name":"neutron"
			} ]
		}
	}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost/identity/v3/auth/tokens",
		httpmock.NewStringResponder(200, body))

	httpmock.RegisterResponder("GET", "http://localhost/compute/v2.1/flavors/2/os-extra_specs",
		httpmock.NewStringResponder(200, `{"extra_specs": {"hw:cpu_policy": "shared", "rmd:swift_workload_policy_url": "RmdPolicies/custom.json", "rmd:glance_workload_policy_uuid": "df1a5223-e297-40d9-9a58-32387a58bed7", "hw:cpu_thread_policy": "prefer"}}`))

	type args struct {
		flavorID                string
		useSwiftInsteadOfGlance bool
	}
	tests := []struct {
		name                string
		args                args
		want                string
		endpointExistsInMap bool
	}{
		{"Correct Glance case", args{"2", false}, "df1a5223-e297-40d9-9a58-32387a58bed7", true},
		{"Correct Swift case", args{"2", true}, "RmdPolicies/custom.json", true},
		{"Lack of Nova endpoint in map case", args{"2", false}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.endpointExistsInMap == false {
				//for this test case we want to remove nova from endpointsMap
				_, ok := endpointsMap["nova"]
				if ok {
					delete(endpointsMap, "nova")
				}
			}

			if got := getWorkloadPolicyUUID(tt.args.flavorID, tt.args.useSwiftInsteadOfGlance); got != tt.want {
				t.Errorf("getWorkloadPolicyUUID() = %v, want %v", got, tt.want)
			}

			if tt.endpointExistsInMap == false {
				//restore nova in endpointsMap
				endpointsMap["nova"] = "http://localhost/compute/v2.1"
			}

		})
	}
}

func Test_getSwiftWorkloadPolicyByURL(t *testing.T) {

	body := `
	{
		"token":{

		"catalog":[
			 ]
		}
	}`

	testWorkload := &workload.RDTWorkLoad{
		Policy: "gold",
	}

	var cachesValue uint32 = 4
	pStateRatio := 3.0
	pStateMonitoring := "on"

	testWorkload2 := &workload.RDTWorkLoad{}
	testWorkload2.Rdt.Cache.Max = &cachesValue
	testWorkload2.Rdt.Cache.Min = &cachesValue
	testWorkload2.PState.Ratio = &pStateRatio
	testWorkload2.PState.Monitoring = &pStateMonitoring

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://10.237.214.102/identity/v3/auth/tokens",
		httpmock.NewStringResponder(200, body))

	httpmock.RegisterResponder("GET", "http://10.237.214.102:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef/RmdPolicies/custom.json",
		httpmock.NewStringResponder(200, `{"policy":"gold"}`))

	httpmock.RegisterResponder("GET", "http://10.237.214.102:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef/RmdPolicies/custom2.json",
		httpmock.NewStringResponder(200, `{"Cache":{"Max":4,"Min":4},"PState":{"Ratio":3.0,"Monitoring":"on"}}`))

	httpmock.RegisterResponder("GET", "http://10.237.214.102:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef/RmdPolicies/error.json",
		httpmock.NewStringResponder(200, `{"Cache":{"Max":4,"Min":`))

	type args struct {
		workloadPolicyURL string
	}
	tests := []struct {
		name                string
		args                args
		want                *workload.RDTWorkLoad
		wantErr             bool
		endpointExistsInMap bool
	}{
		{"Correct Gold Policy case", args{"RmdPolicies/custom.json"}, testWorkload, false, true},
		{"Correct Cache and PState Policy case", args{"RmdPolicies/custom2.json"}, testWorkload2, false, true},
		{"Lack of Swift endpoint in map case", args{"RmdPolicies/custom.json"}, nil, true, false},
		{"Malformed data case", args{"RmdPolicies/error.json"}, nil, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.endpointExistsInMap == false {
				//for this test case we want to remove nova from endpointsMap
				_, ok := endpointsMap["swift"]
				if ok {
					delete(endpointsMap, "swift")
				}
			}

			got, err := getSwiftWorkloadPolicyByURL(tt.args.workloadPolicyURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSwiftWorkloadPolicyByURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSwiftWorkloadPolicyByURL() = %v, want %v", got, tt.want)
			}

			if tt.endpointExistsInMap == false {
				//restore swift in endpointsMap
				endpointsMap["swift"] = "http://10.237.214.102:8080/v1/AUTH_0ff337bec7d4415dbc60ffca09db14ef"
			}
		})
	}
}
