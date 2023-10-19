package handler

import (
	"fmt"
	"sync"
)

// TaskMap 任务 Map
var TaskMap = NewDockerPullTaskMap()

func NewDockerPullTaskMap() *dockerPullTaskMap {
	return &dockerPullTaskMap{
		Maps: make(map[string]*DockerPullTask, 0),
		Lock: sync.Mutex{},
	}
}

// dockerPullTaskMap docker pull 任务 map
type dockerPullTaskMap struct {
	Maps map[string]*DockerPullTask
	Lock sync.Mutex
}

func (d *dockerPullTaskMap) Add(taskId string, task *DockerPullTask) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	d.Maps[taskId] = task
}

func (d *dockerPullTaskMap) Del(taskId string) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	delete(d.Maps, taskId)
}

func (d *dockerPullTaskMap) Get(taskId string) (*DockerPullTask, bool) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	task, ok := d.Maps[taskId]
	return task, ok
}

// SetStateRun 设置正在进行的状态
func (d *dockerPullTaskMap) SetStateRun(taskId string) error {
	task, ok := d.Get(taskId)
	if !ok {
		return fmt.Errorf(taskId + " is null")
	}
	task.State = 1
	return nil
}

// SetStateComplete 设置完成的状态
func (d *dockerPullTaskMap) SetStateComplete(taskId string) error {
	task, ok := d.Get(taskId)
	if !ok {
		return fmt.Errorf(taskId + " is null")
	}
	task.State = 2
	return nil
}

// SetStateBreakOff 设置中断的状态
func (d *dockerPullTaskMap) SetStateBreakOff(taskId string) error {
	task, ok := d.Get(taskId)
	if !ok {
		return fmt.Errorf(taskId + " is null")
	}
	task.State = 3
	return nil
}

// SetResult 设置结果
func (d *dockerPullTaskMap) SetResult(taskId string, b interface{}) error {
	task, ok := d.Get(taskId)
	if !ok {
		return fmt.Errorf(taskId + " is null")
	}
	task.Result = b
	return nil
}

type DockerPullTask struct {
	IP     string
	Result interface{}
	State  int // 0: 预留， 1: 进行中， 2: 完成, 3: 中断
}

func (d *DockerPullTask) GetResult() *DockerPullTask {
	return d
}
