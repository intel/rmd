package testhelpers

import (
	"fmt"
	"os"
	"testing"
)

func TestCreateNewProcess(t *testing.T) {
	p, err := CreateNewProcess("sleep 10")
	if err != nil {
		t.Errorf("error %v", err)
	}

	fmt.Println(p.Pid)
	err = p.Kill()
	if err != nil {
		t.Errorf("error %v", err)
	}

}

func TestAssembleRequest(t *testing.T) {
	p, _ := CreateNewProcess("sleep 10")
	fmt.Println(AssembleRequest([]*os.Process{p}, []string{}, 10, 10, ""))
	CleanupProcess(p)
}

func TestAssembleRequestMultipleProcess(t *testing.T) {
	ps, _ := CreateNewProcesses("sleep 10", 3)
	fmt.Println(AssembleRequest(ps, []string{}, 10, 10, ""))
	CleanupProcesses(ps)
}

func TestAssembleRequestCPUs(t *testing.T) {
	fmt.Println(AssembleRequest([]*os.Process{}, []string{"1-2"}, 10, 10, ""))
}

func TestFormatByKey(t *testing.T) {
	m := map[string]interface{}{"w": "world"}
	s := "Hello {{.w}}!"
	r, _ := FormatByKey(s, m)
	if r != "Hello world!" {
		t.Errorf("error for format string: %s", s)
	}

	m = map[string]interface{}{"age": "47", "name": "John", "key": "no use"}
	s = "Hi {{.name}}. Your age is {{.age}}\n"
	r, _ = FormatByKey(s, m)
	if r != "Hi John. Your age is 47\n" {
		t.Errorf("error for format string: %s", s)
	}
}
