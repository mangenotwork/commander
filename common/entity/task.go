package entity

// Task 任务
type Task struct {
	ID       string
	IP       string
	Result   interface{} // 任务日志等等
	State    int         // 1: 进行中， 2: 完成, 3: 中断
	StateStr string
	Create   string // 任务创建时间
	Note     string // 任务描述
}

// Operate  操作记录
type Operate struct {
	Date     string
	ClientIp string
	Note     string
}
