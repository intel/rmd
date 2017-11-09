package resctrl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/intel/rmd/util"
	"github.com/intel/rmd/util/task"
	log "github.com/sirupsen/logrus"
)

// Task is task for resctrl
type Task struct {
	TaskName string
	*ResAssociation
	RessSnapshot map[string]*ResAssociation
	Group        string
	Path         string
	Revert       bool // whether need to Revert after task faild
}

// Name returns name of the task
func (t Task) Name() string {
	return t.TaskName
}

// Run starts the task
func (t Task) Run() error {
	return nil
}

// Rollback task
func (t Task) Rollback() error {
	return nil
}

// GroupTask is task to create new group
type GroupTask struct {
	Task
}

// Run to create new resource group
func (t GroupTask) Run() error {
	return os.MkdirAll(t.Path, 0755)
}

// Rollback remove created resrouce group
func (t GroupTask) Rollback() error {
	os.Remove(t.Path)
	return nil
}

// CPUsTask is the CPU task
type CPUsTask struct {
	Task
}

// Run to write CPU mask
func (t CPUsTask) Run() error {
	// Only write to cpus if admin specify cpu bit map
	// only commit a user deinfed cpus
	if t.CPUs != "" {
		return writeFile(t.Path, "cpus", t.CPUs)
	}
	// NOTE: CPUS is "" means no need to change the cpus file.
	return nil
}

// Rollback dos nothing for now
func (t CPUsTask) Rollback() error {
	if !t.Revert {
		return nil
	}
	// Well, when Roll back cpus in resctrl, write back it to previous value
	// if applicable, there's no need to touch other group's cpus, as
	// workload layer will make sure a commit have no overlap with others
	res, ok := t.RessSnapshot[t.Group]
	if ok {
		return writeFile(t.Path, "cpus", res.CPUs)
	}
	return nil
}

// TasksTask is the task for add tasks
type TasksTask struct {
	Task
}

// Run add tasks
func (t TasksTask) Run() error {
	// only commit a user deinfed group's task to sys fs
	if t.Group != "." && len(t.Tasks) > 0 {
		// write one task one time, or write will fail
		for _, v := range t.Tasks {
			err := writeFile(t.Path, "tasks", v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Rollback tasks
func (t TasksTask) Rollback() error {
	if !t.Revert {
		return nil
	}

	res, ok := t.RessSnapshot[t.Group]
	if ok {
		var ts []string
		// Find if origin tasks list has more tasks than the new commit
		// when do roll back, try to write back task to the origin group
		ts = util.SubtractStringSlice(res.Tasks, t.ResAssociation.Tasks)
		if len(ts) > 0 {
			// Write back to origin
			for _, v := range ts {
				err := writeFile(t.Path, "tasks", v)
				if err != nil {
					log.Warningf("Failed to add back task %s wile try to rollbak task for %s", v, t.Group)
				}
			}
		} else {
			// Find if new commit task has more tasks than origin one.
			// Remove them to default group
			ts = util.SubtractStringSlice(t.ResAssociation.Tasks, res.Tasks)
			if err := RemoveTasks(ts); err != nil {
				log.Warnln(err)
			}
		}
	}
	return nil
}

// SchemataTask is the task to create schemata
type SchemataTask struct {
	Task
}

// Run to commit schemata
func (t SchemataTask) Run() error {
	if len(t.Schemata) > 0 {
		schemata := make([]string, 0, 10)
		for k, v := range t.Schemata {
			str := make([]string, 0, 10)
			// resctrl require we have strict cache id order
			for cacheid := 0; cacheid < len(v); cacheid++ {
				for _, cos := range v {
					if uint8(cacheid) == cos.ID {
						str = append(str, fmt.Sprintf("%d=%s", cos.ID, cos.Mask))
						break
					}
				}
			}
			schemata = append(schemata, strings.Join([]string{k, strings.Join(str, ";")}, ":"))
		}
		data := strings.Join(schemata, "\n")
		err := writeFile(t.Path, "schemata", data)
		return err
	}
	return nil
}

// Rollback to revert it
func (t SchemataTask) Rollback() error {
	// NOTE, do not need to revert the Schemata to the snapshort
	return nil
}

func taskFlow(group string, r *ResAssociation, rs map[string]*ResAssociation) error {
	tasks := []task.Task{}
	path := SysResctrl

	if strings.ToLower(group) != "default" && group != "." {
		path = filepath.Join(SysResctrl, group)
	}

	ct := CPUsTask{Task{"update-cpus", r, rs, group, path, true}}
	tt := TasksTask{Task{"update-tasks", r, rs, group, path, true}}
	st := SchemataTask{Task{"update-schemata", r, rs, group, path, true}}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		gt := GroupTask{Task{"creat-group", r, rs, group, path, true}}
		ct.Revert = false
		tt.Revert = false
		st.Revert = false
		tasks = append(tasks, gt)
	}

	tasks = append(tasks, []task.Task{ct, tt, st}...)
	taskList := task.NewTaskList(tasks)
	if err := taskList.Start(); err != nil {
		log.Errorf("Failed to execute task list %s", err.Error())
		return err
	}
	return nil
}
