package handler

import (
	"fmt"
	"sync"
)

// DockerRunTaskMap docker运行任务map
var DockerRunTaskMap = NewDockerRunTaskMap()

func NewDockerRunTaskMap() *dockerRunTaskMap {
	return &dockerRunTaskMap{
		Maps : make(map[string]*DockerRunTask, 0),
		Lock : sync.Mutex{},
	}
}

type dockerRunTaskMap struct {
	Maps map[string]*DockerRunTask
	Lock sync.Mutex
}

func (d *dockerRunTaskMap) Add(taskId string, task *DockerRunTask){
	d.Lock.Lock()
	defer d.Lock.Unlock()
	d.Maps[taskId] = task
}

func (d *dockerRunTaskMap) Del(taskId string) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	delete(d.Maps, taskId)
}

func (d *dockerRunTaskMap) Get(taskId string) (*DockerRunTask, bool){
	d.Lock.Lock()
	defer d.Lock.Unlock()
	task, ok := d.Maps[taskId]
	return task, ok
}

// SetStateRun 设置正在进行的状态
func (d *dockerRunTaskMap) SetStateRun(taskId string) error {
	task, ok := d.Get(taskId)
	if !ok {
		return fmt.Errorf(taskId+" is null")
	}
	task.State = 1
	return nil
}

// SetStateComplete 设置完成的状态
func (d *dockerRunTaskMap) SetStateComplete(taskId string) error {
	task, ok := d.Get(taskId)
	if !ok {
		return fmt.Errorf(taskId+" is null")
	}
	task.State = 2
	return nil
}

// SetStateBreakOff 设置中断的状态
func (d *dockerRunTaskMap) SetStateBreakOff(taskId string) error {
	task, ok := d.Get(taskId)
	if !ok {
		return fmt.Errorf(taskId+" is null")
	}
	task.State = 3
	return nil
}

// SetResult 设置结果
func (d *dockerRunTaskMap) SetResult(taskId, b string) error {
	task, ok := d.Get(taskId)
	if !ok {
		return fmt.Errorf(taskId+" is null")
	}
	task.Result = b
	return nil
}

type DockerRunTask struct {
	IP string
	Result string
	State int  // 0: 预留， 1: 进行中， 2: 完成, 3: 中断
}

func (d *DockerRunTask) GetResult() *DockerRunTask {
	return d
}
