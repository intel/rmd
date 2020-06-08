package testhelpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"text/template"

	"github.com/spf13/viper"
)

// CreateNewProcess Create a new background running process by command string
// e.g.
// CreateNewprocess("sleep 1000")
func CreateNewProcess(cmd string) (*os.Process, error) {
	cmdCmd := exec.Command("bash", "-c", cmd)
	err := cmdCmd.Start()
	if err != nil {
		return nil, err
	}
	return cmdCmd.Process, nil
}

// CreateNewProcesses To create some processes
func CreateNewProcesses(cmd string, number int) ([]*os.Process, error) {
	var ps []*os.Process
	for i := 0; i < number; i++ {
		p, err := CreateNewProcess(cmd)
		if err != nil {
			return ps, err
		}
		ps = append(ps, p)
	}
	return ps, nil
}

// CleanupProcess To kill process
func CleanupProcess(p *os.Process) {
	p.Kill()
}

// CleanupProcesses To kill processes
func CleanupProcesses(ps []*os.Process) {
	for _, p := range ps {
		p.Kill()
	}
}

// AssembleRequest assemble the request body by given process id and max, min cache or policy
func AssembleRequest(processes []*os.Process, coreIds []string, maxCache, minCache int, mbaPercentage int, policy string) map[string]interface{} {
	data := map[string]interface{}{}
	if policy != "" {
		data["policy"] = policy
	} else {
		params := map[string]int{"max": maxCache, "min": minCache}

		if mbaPercentage != -1 {
			params1 := map[string]int{"percentage": mbaPercentage}
			data["rdt"] = map[string]interface{}{"cache": params, "mba": params1}
		} else {
			data["rdt"] = map[string]interface{}{"cache": params}
		}
	}

	var taskIds []string
	for _, p := range processes {
		pStr := strconv.Itoa(p.Pid)
		taskIds = append(taskIds, pStr)
	}
	if len(taskIds) > 0 {
		data["task_ids"] = taskIds
	}
	if len(coreIds) > 0 {
		data["core_ids"] = coreIds
	}

	return data
}

// ConfigInit initializes config
func ConfigInit(path string) error {
	viper.SetConfigType("toml")
	viper.SetConfigName("rmd")  // name of config file (without extension)
	viper.AddConfigPath("/tmp") // path to look for the config file in
	err := viper.ReadInConfig() // Find and read the config file
	return err
}

// GetConfigOptions is just simple wraper for config Unmarshal
func GetConfigOptions(rawVal interface{}) error {
	return viper.Unmarshal(rawVal)
}

// GetConfigOption is just simple wraper for config UnmarshalKey
func GetConfigOption(key string, rawVal interface{}) error {
	return viper.UnmarshalKey(key, rawVal)
}

//GetConfigDebugPort getter for Debug port
func GetConfigDebugPort() int {
	var port int
	GetConfigOption("debug.debugport", &port)
	return port
}

//GetConfigTLSPort getter for TLS port
func GetConfigTLSPort() int {
	var port int
	GetConfigOption("default.tlsport", &port)
	return port
}

//GetConfigAddr getter for address
func GetConfigAddr() string {
	var addr string
	GetConfigOption("default.address", &addr)
	return addr
}

//GetHTTPV1URL getter for HTTPV1URL
func GetHTTPV1URL() string {
	return fmt.Sprintf(
		"http://%s:%d/v1/", GetConfigAddr(), GetConfigDebugPort())
}

//GetHTTPSV1URL getter for HTTPSV1URL
func GetHTTPSV1URL() string {
	return fmt.Sprintf(
		"https://%s:%d/v1/", GetConfigAddr(), GetConfigTLSPort())
}

//GetClientAuthType getter for client auth type
func GetClientAuthType() string {
	var clientAuthTypePath string
	GetConfigOption("default.clientauth", &clientAuthTypePath)
	return clientAuthTypePath
}

//GetPolicyPath getter for path with policy
func GetPolicyPath() string {
	var policyPath string
	GetConfigOption("default.policypath", &policyPath)
	return policyPath
}

// FormatByKey usage:
//       m := map[string]interface{}{"name": "John", "age": 47}
//       s := "Hi {{.name}}. Your age is {{.age}}\n"
//		 result, err := FormatByKey(s, m)
// use go native template.
// If you do not like go template, you can use mustache
//       import "github.com/hoisie/mustache"
//       m := map[string]interface{}{"name": "John", "age": 47}
//       s := "Hi {{name}}. Your age is {{age}}\n"
//       result := mustache.Render("hello {{c}}, {{age}}", m)
func FormatByKey(f string, m map[string]interface{}) (string, error) {
	var tpl bytes.Buffer
	t := template.Must(template.New("").Parse(f))
	if err := t.Execute(&tpl, m); err != nil {
		return f, err
	}
	return tpl.String(), nil
}
