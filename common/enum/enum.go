package enum

const SlaveVersion = "v0.3"

const (
	TaskStateRun      TaskState = 1
	TaskStateComplete TaskState = 2
	TaskStateBreakOff TaskState = 3
)

// TaskState 任务状态
type TaskState int

func (t TaskState) Str() string {
	return TaskStateMap[t]
}

func (t TaskState) Value() int {
	return int(t)
}

var TaskStateMap = map[TaskState]string{
	TaskStateRun:      "任务执行中",
	TaskStateComplete: "任务执行完成",
	TaskStateBreakOff: "任务执行被中断",
}

// ExecutableStateExecuting 任务执行中标识
const ExecutableStateExecuting = "executing"

// ExecutableStateDiscontinued 任务执行停止标识
const ExecutableStateDiscontinued = "discontinued"
