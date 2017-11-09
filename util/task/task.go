package task

import (
	"container/list"
)

// Task is an interface which could be added into a TaskList
type Task interface {
	// Name of the task
	Name() string
	// Run method
	Run() error
	// Rollback if failed to run this task
	Rollback() error
}

// List is a list which could be called as transaction
type List struct {
	taskList *list.List
}

// NewTaskList initials a list of task
func NewTaskList(tasks []Task) *List {
	tl := list.New()
	for _, t := range tasks {
		tl.PushBack(t)
	}
	return &List{taskList: tl}
}

// Start to run this tasks list, will return the last run error
func (l *List) Start() error {
	var fe *list.Element
	var runErr error

	for te := l.taskList.Front(); te != nil; te = te.Next() {
		t := te.Value.(Task)
		if err := t.Run(); err != nil {
			fe = te
			runErr = err
			break
		}
	}

	if runErr != nil {
		// Rollback
		for te := fe; te != nil; te = te.Prev() {
			t := te.Value.(Task)
			if err := t.Rollback(); err != nil {
				// Save rollback error and break?
			}
		}
		return runErr
	}

	return nil
}
