package main

import (
	"encoding/json"
	"flag" // flag is enough for us.
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/intel/rmd/cmd/template"
	"github.com/intel/rmd/lib/cpu"
	"github.com/intel/rmd/test/test_helpers" // TODO: should move it to source path
)

var gopherType string

const (
	confPath    = "./rmd.toml"
	defPlatform = "Broadwell"
)

func genDefaultPlatForm() string {
	dpf := strings.Title(cpu.GetMicroArch(cpu.GetSignature()))
	if dpf == "" {
		fmt.Println("Do not support the host platform, use defaut platform:", defPlatform)
		dpf = defPlatform
	}
	return dpf
}

func genPlatFormMap() map[string]bool {
	m := cpu.NewCPUMap()
	pfm := map[string]bool{}
	for _, value := range m {
		pfm[strings.Title(value)] = true
	}
	return pfm
}

func genPlatFormList(pfm map[string]bool) []string {
	pfs := []string{}
	for k := range pfm {
		pfs = append(pfs, k)
	}
	return pfs
}

func mergeOptions(options ...map[string]interface{}) map[string]interface{} {
	union := make(map[string]interface{})
	for _, o := range options {
		for k, v := range o {
			union[k] = v
		}
	}
	return union
}

func main() {
	path := flag.String("path", confPath, "the path of the generated rmd config.")

	dpf := genDefaultPlatForm()
	pfm := genPlatFormMap()
	pfs := genPlatFormList(pfm)
	platform := flag.String("platform", dpf,
		"the platform than rmd will run, Support PlatForm:\n\t    "+strings.Join(pfs, ", "))
	datas := flag.String("data", "",
		"Data options that overwrite the config opitons. "+
			"It can be a json format as '{\"a\": 1, \"b\": 2}' or "+
			"a key/value assignment format as \"key=value,key=value\". "+
			"It can also start a letter @, the rest should be a file name to read the json data from.")
	flag.Parse()

	if _, ok := pfm[*platform]; !ok {
		fmt.Println("Error, unsupport platform:", *platform)
		os.Exit(1)
	}

	union := []map[string]interface{}{template.Options}

	//  Skylake, Kaby Lake, Broadwell
	// FIXME hard code, a smart way to load the platform and other optionts automatically.
	if *platform == "Broadwell" {
		union = append(union, template.Broadwell)
	}
	if *platform == "Skylake" {
		union = append(union, template.Skylake)
	}

	var bdatas []byte
	if strings.HasPrefix(*datas, "@") {
		f := strings.TrimPrefix(*datas, "@")
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			fmt.Println("Bad data file: ", err)
			os.Exit(1)
		}
		bdatas = dat
	} else if strings.HasPrefix(*datas, "{") {
		bdatas = []byte(*datas)
	} else if strings.Contains(*datas, "=") {
		// TODO will also  support -data 'stdout=true,tasks=["ovs*",dpdk]'
		fmt.Println("Unsupport key/value assignment format at present.")
		os.Exit(1)
	}
	if len(bdatas) > 2 {
		var js interface{}
		if err := json.Unmarshal([]byte(bdatas), &js); err != nil {
			fmt.Println("Error to paser data:", err)
			os.Exit(1)
		}
		union = append(union, js.(map[string]interface{}))
	}

	option := mergeOptions(union...)
	conf, err := testhelpers.FormatByKey(template.Templ, option)
	if err != nil {
		fmt.Println("Error, to generate config file:", err)
		os.Exit(1)
	}

	f, err := os.Create(*path)
	if err != nil {
		fmt.Println("Error, to create config file:", err)
		os.Exit(1)
	}
	defer f.Close()
	_, err = f.WriteString(conf)
	if err != nil {
		fmt.Println("Error, to write config file:", err)
		os.Exit(1)
	}
	f.Sync()
}
