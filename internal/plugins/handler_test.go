package plugins

import (
	"errors"
	"testing"

	"github.com/emicklei/go-restful"
)

type FakeModule struct {
}

func (fm *FakeModule) Initialize(params map[string]interface{}) error {
	return nil
}

func (fm *FakeModule) GetEndpointPrefixes() []string {
	return []string{"ep1", "ep2"}
}

func (fm *FakeModule) HandleRequest(request *restful.Request, response *restful.Response) {
}

func (fm *FakeModule) Validate(params map[string]interface{}) error {
	if len(params) == 0 {
		return errors.New("No params")
	}
	return nil
}

func (fm *FakeModule) Enforce(params map[string]interface{}) (string, error) {
	if len(params) == 0 {
		return "", errors.New("No params")
	}
	return "111", nil
}

func (fm *FakeModule) Release(params map[string]interface{}) error {
	if len(params) == 0 {
		return errors.New("No params")
	}
	return nil
}

func (fm *FakeModule) GetCapabilities() string {
	return ""
}

var goodParams = map[string]interface{}{"param1": "text", "param2": 123}
var badParams = map[string]interface{}{}

func init() {
	var fm FakeModule
	Interfaces["mod1"] = nil
	Interfaces["mod2"] = &fm
}

func TestEnforce(t *testing.T) {
	type args struct {
		moduleName string
		params     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Invalid module name and bad params", args{"modX", badParams}, "", true},
		{"Invalid module and bad params", args{"mod1", badParams}, "", true},
		{"Good module and bad params", args{"mod2", badParams}, "", true},
		{"Good module and good params", args{"mod2", goodParams}, "111", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Enforce(tt.args.moduleName, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enforce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Enforce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelease(t *testing.T) {
	type args struct {
		moduleName string
		params     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Invalid module name and bad params", args{"modX", badParams}, true},
		{"Invalid module and bad params", args{"mod1", badParams}, true},
		{"Good module and bad params", args{"mod2", badParams}, true},
		{"Good module and good params", args{"mod2", goodParams}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Release(tt.args.moduleName, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("Release() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	type args struct {
		moduleName string
		params     map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Invalid module name and bad params", args{"modX", badParams}, true},
		{"Invalid module and bad params", args{"mod1", badParams}, true},
		{"Good module and bad params", args{"mod2", badParams}, true},
		{"Good module and good params", args{"mod2", goodParams}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.moduleName, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore(t *testing.T) {
	type args struct {
		name  string
		iface ModuleInterface
	}
	var fm FakeModule
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Empty name", args{"", &fm}, true},
		{"NIL interface", args{"store1", nil}, true},
		{"Correct case", args{"store2", &fm}, false},
		{"Repeated name", args{"store2", &fm}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Store(tt.args.name, tt.args.iface); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
