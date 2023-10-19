package entity

// ExecutableFile 可执行文件
type ExecutableFile struct {
	Name string
	FileName string
	SaveFilePath string // 保存到磁盘的路径
	Path string  // 下载的路径
	DirPath string //  解压的目录路径
	Size string
	OSType string
	UploadTime string
	Md5 string
	FileID string // 文件名称md5
	Cmd string // 可执行文件运行的命令
	Env []string // 可执行文件运行的环境变量
}

// ExecutableDeployArg 部署可执行文件
type ExecutableDeployArg struct {
	Slave string  // slave 地址 ip
	DownloadFile string
	Arg string
	TaskId string
	Note string
	Cmd string
	Env string
}

// ExecutableDeployTask 可执行文件部署任务
type ExecutableDeployTask struct {
	Slave string  // slave 地址 ip
	TaskId string   // 任務id
	Command string   // 執行的全部命令
	Env string
	ExecutableName string   // 執行文件名  對應  ExecutableFile.Name
	Arg string   // 執行的參數
	Time string
	PID int
	State string // 狀態
	Note string // 備註
}

// ExecutableKillArg 可执行文件kill
type ExecutableKillArg struct {
	PID string
	Value string
	TaskId string
}

// ExecutableKillRse 可执行文件kill
type ExecutableKillRse struct {
	Rse string
	TaskId string
}

// ExecutableProcess 可执行文件进程
type ExecutableProcess struct {
	Slave string
	Project string
	PID string
	TaskID string
	Cmd string
}

// ExecutableRunStateArg 可执行文件运行
type ExecutableRunStateArg struct {
	PId string
	TaskId string
}

// ExecutableRunStateRse 可执行文件运行
type ExecutableRunStateRse struct {
	PId string
	TaskId string
	Rse string
}
