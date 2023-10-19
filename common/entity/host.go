package entity

// HostInfo host信息
type HostInfo struct {
	SlaveVersion string `json:"slave_version"`

	// salve的ip地址
	Slave string `json:"ip"`

	//是否在线
	SlaveOnline string `json:"online"`
	HostName    string `json:"host_name"`

	//系统平台
	SysType string `json:"sys_type"`

	//系统版本 os_name+版号
	OsName string `json:"os_name"`
	OsNum  string `json:"os_num"`

	//系统架构
	SysArchitecture string `json:"sys_architecture"`

	//CPU核心数
	CpuCoreNumber string `json:"cpu_core_number"`

	//CPU name
	CpuName string `json:"cpu_name"`

	//CPU ID
	CpuID string `json:"cpu_id"`

	//主板ID
	BaseBoardID string `json:"board_id"`

	//内存总大小 MB
	MemTotal string `json:"mem_totle"`

	//磁盘信息
	Disk []*DiskInfo `json:"disk"`

	//磁盘总大小 MB
	DiskTotal string `json:"disk_totle"`

	// 当前CPU使用率
	NowCPURate string `json:"now_cpu_rate"`

	// 当前内存使用率
	NowMEMRate string `json:"now_mem_rate"`

	// 是否安装Docker
	HasDocker string `json:"has_docker"`

	// docker 版本
	DockerVersion string `json:"docker_version"`

	// 第一个数字表示系统运行时间
	RunTime string `json:"run_time"`
	// 第二个数字表示系统空闲时间
	LdleTime string `json:"ldle_time"`

	// slave的文件服务
	FileServerPort   string `json:"file_server_port"`
	FileServerSecret string `json:"file_server_secret"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	DiskName    string
	DistType    string
	DistTotalMB string
	DistUse     *DiskUseInfo
}

// DiskUseInfo 磁盘使用的信息
type DiskUseInfo struct {
	Total int     //MB
	Free  int     //MB
	Rate  float32 //%
}

// MemInfo 内存信息
type MemInfo struct {
	//所有可用RAM大小 （即物理内存减去一些预留位和内核的二进制代码大小）
	MemTotal int64 `json:"mem_total"`

	//内存使用
	MemUsed int64 `json:"mem_used"`

	//LowFree与HighFree的总和，被系统留着未使用的内存
	MemFree int64 `json:"mem_free"`

	//用来给文件做缓冲大小
	MemBuffers int64 `json:"mem_buffers"`

	//被高速缓冲存储器（cache memory）用的内存的大小（等于diskcache minus SwapCache ）.
	MemCached int64 `json:"mem_cached"`
}

// NetWorkIO 网络IO
// 单位 (kb/sec)
type NetWorkIO struct {
	Name string
	Tx   float32 //发送
	Rx   float32 //接收
}

// ProcMemInfo 从 proc/meminfo 获取内存信息
type ProcMemInfo struct {
	//所有可用RAM大小 （即物理内存减去一些预留位和内核的二进制代码大小）
	MemTotal int64 `json:"mem_total"`

	//内存使用
	MemUsed int64 `json:"mem_used"`

	//LowFree与HighFree的总和，被系统留着未使用的内存
	MemFree int64 `json:"mem_free"`

	//用来给文件做缓冲大小
	MemBuffers int64 `json:"mem_buffers"`

	//被高速缓冲存储器（cache memory）用的内存的大小（等于diskcache minus SwapCache ）.
	MemCached int64 `json:"mem_cached"`
}

// CPUUseRate cpu使用率
type CPUUseRate struct {
	CPU     string
	UseRate float32
}

// SlavePerformance 心跳包上传健康信息
type SlavePerformance struct {
	// CPU使用率
	CPU     *CPUUseRate
	CPUCore []*CPUUseRate

	// 内存使用率
	MEM *ProcMemInfo

	// 磁盘使用率
	Disk []*DiskInfo

	// 网络IO
	NetWork []*NetWorkIO

	// 连接数
	ConnectNum int

	TimeStamp string
}

// ProcessBaseInfo 进程基本信息
// 是兼容性的结构体 兼容了linux 和 windows
type ProcessBaseInfo struct {
	PID   string `json:"pid"`
	User  string `json:"user"`
	PName string `json:"pname"`
	PPID  string `json:"ppid"` //父进程
	C     string
	Stime string
	TTY   string `json:"tty"`
	Time  string
	CMD   string `json:"cmd"` //执行命令
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string
	Size    int64
	Mode    int         // 文件模式位
	ModTime string      // 修改时间
	IsDir   bool        // 是否目录
	Sys     interface{} // 基础数据源（可以返回nil）
}

// ProcessArg 执行进程的参数
type ProcessArg struct {
	Name   string
	Type   int // 1:宿主机上的命令   2:可执行的二进制
	TaskId string
	Cmd    string
	Arg    []string
}

// PortInfo 端口信息
type PortInfo struct {
	ProtoType      string
	LocalAddress   string
	ForeignAddress string
	State          string
	PID            string
	PName          string
}

// ProcessKillArg 关闭進程的參數
type ProcessKillArg struct {
	PID   string
	Value string
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PCPU      float64
	Cmd       string
	Environ   []string
	StatusTxt string // 状态描述
}

// MonitorRule 监控规则
type MonitorRule struct {
	Slave         string
	MaxCPU        int // 允许最大cpu使用率，超过则报警
	MaxMem        int // 允许最大内存使用率，超过则报警
	MaxDisk       int // 允许最大磁盘使用率，超过则报警
	MaxTx         int // 允许最大网络TX，超过则报警
	MaxRx         int // 允许最大网络RX，超过则报警
	MaxConnectNum int // 允许最大网络连接数，超过则报警
}

// 监控规则 默认值， 没有设置则取默认值
var (
	MonitorRuleMaxCPUDefault        int = 30
	MonitorRuleMaxMemDefault        int = 60
	MonitorRuleMaxDiskDefault       int = 60
	MonitorRuleMaxTxDefault         int = 10
	MonitorRuleMaxRxDefault         int = 10
	MonitorRuleMaxConnectNumDefault int = 100
)

// Alarm 报警
type Alarm struct {
	ID    string
	Slave string
	Date  string
	Note  string
	Lv    string
}

// SlaveHostsRse hosts文件
type SlaveHostsRse struct {
	Data string
}

// SlaveHostsArg 修改hosts文件
type SlaveHostsArg struct {
	TaskId string
	Data   string
}

// EnvDeployedCheckArg 判断是否安装这些软件
type EnvDeployedCheckArg struct {
	Software []string // 软件名支持多个  如: docker;nginx;
}

type EnvDeployedCheckRse struct {
	SoftwareCheck []*SoftwareDeployedCheck
}

// SoftwareDeployedCheck 判断部署的软件
type SoftwareDeployedCheck struct {
	Software string
	IsHave   bool
	Info     string
}

// InstallDockerRse 安装docker 返回信息
type InstallDockerRse struct {
	Rse string
}

// RemoveDockerRse 卸载docker 返回信息
type RemoveDockerRse struct {
	Rse string
}

// InstallNginxRse 安装nginx 返回信息
type InstallNginxRse struct {
	Rse string
}

// RemoveNginxRse 卸载nginx 返回信息
type RemoveNginxRse struct {
	Rse string
}

// ProxyHttpCreateArg 部署一个http/s代理
type ProxyHttpCreateArg struct {
	Name   string
	Slave  string
	Port   string
	Note   string
	TaskId string
}

type ProxyHttpCreateRse struct {
	Name   string
	Slave  string
	Port   string
	Note   string
	TaskId string
	Error  string
	Rse    string
}

// HttpsProxy http/s 代理数据结构
type HttpsProxy struct {
	Name    string
	Slave   string
	Port    string
	Create  string
	IsClose string // 是否关闭  0 否  1 是
	Note    string // 备注
	IsDel   string // 是否删除  0 否 1 是
}

// ProxyHttpUpdateArg http/s 代理修改
type ProxyHttpUpdateArg struct {
	Name       string // 代理名称
	UpdateType string // 修改的字段名
	Vlaue      string // 修改的值
	TaskId     string
}

type ProxyHttpUpdateRse struct {
	Name   string // 代理名称
	Rse    string // 修改结果
	TaskId string
}

// Socket5Proxy socket5 代理数据结构
type Socket5Proxy struct {
	Name    string
	Slave   string
	Port    string
	Create  string
	IsClose string // 是否关闭  0 否  1 是
	Note    string // 备注
	IsDel   string // 是否删除  0 否 1 是
}

// Socket5ProxyCreateArg  部署socket5代理
type Socket5ProxyCreateArg struct {
	Name   string
	Slave  string
	Port   string
	Note   string
	TaskId string
}

type Socket5ProxyCreateRse struct {
	Name   string
	Slave  string
	Port   string
	Note   string
	TaskId string
	Error  string
	Rse    string
}

// ProxySocket5UpdateArg http/s 代理修改
type ProxySocket5UpdateArg struct {
	Name       string // 代理名称
	UpdateType string // 修改的字段名
	Vlaue      string // 修改的值
	TaskId     string
}

type ProxySocket5UpdateRse struct {
	Name   string // 代理名称
	Rse    string // 修改结果
	TaskId string
}

// TCPForward tcp网络转发
type TCPForward struct {
	Name         string
	Slave        string
	Port         string
	Create       string
	Note         string
	ForwardTable []string // 转发表
	IsClose      string   // 是否关闭  0 否  1 是
	IsDel        string   // 是否删除  0 否 1 是
}

// TCPForwardCreateArg tcp 转发 参数
type TCPForwardCreateArg struct {
	Name         string
	Slave        string
	Port         string
	Note         string
	ForwardTable []string // 转发表
	TaskId       string
}

// TCPForwardCreateRse tcp 转发 参数
type TCPForwardCreateRse struct {
	Name         string
	Slave        string
	Port         string
	Note         string
	ForwardTable []string // 转发表
	TaskId       string
	Error        string
	Rse          string
}

// TCPForwardUpdateArg tcp 转发修改
type TCPForwardUpdateArg struct {
	Name       string // 转发名称
	UpdateType string // 修改的字段名
	Vlaue      string // 修改的值
	TaskId     string
}

type TCPForwardUpdateRse struct {
	Name   string // 代理名称
	Rse    string // 修改结果
	TaskId string
}

// GetSlavePathInfoArg 获取slave指定路径的目录结构与文件
type GetSlavePathInfoArg struct {
	Path   string
	TaskId string
}

type GetSlavePathInfoRse struct {
	Error         string
	Rse           string
	TaskId        string
	FileStructure []*FileStructure
}

// FileStructure 文件的状态与属性
type FileStructure struct {
	FileName string
	IsDir    bool
	IsEdit   bool
	IsZip    bool // 是否压缩文件
}

// CatSlaveFileArg 查看slave 文件内容参数
type CatSlaveFileArg struct {
	FilePath string
	TaskId   string
}

type CatSlaveFileRse struct {
	Data   string
	TaskId string
}

// SlaveMkdirArg slave 创建目录
type SlaveMkdirArg struct {
	Path   string
	TaskId string
}

type SlaveMkdirRse struct {
	Rse    string
	TaskId string
}

// SlaveDecompressArg slave 解压文件
type SlaveDecompressArg struct {
	Path   string
	TaskId string
}

type SlaveDecompressRse struct {
	Rse    string
	TaskId string
}

// NginxInfoArg nginx info arg
type NginxInfoArg struct {
	TaskId string
}

type NginxInfo struct {
	Version  string   // 版本
	ConfPath string   // 配置路径
	LogPath  string   // 日志路径
	PID      []string // 进程
	Status   string   // 状态
}

type NginxInfoRse struct {
	TaskId string
	Rse    *NginxInfo
}

// NginxStartArg 启动nginx
type NginxStartArg struct {
	TaskId string
}

type NginxStartRse struct {
	TaskId string
	Rse    string
}

// NginxReloadArg 重启nginx
type NginxReloadArg struct {
	TaskId string
}

type NginxReloadRse struct {
	TaskId string
	Rse    string
}

// NginxQuitArg 停止nginx
type NginxQuitArg struct {
	TaskId string
}

type NginxQuitRse struct {
	TaskId string
	Rse    string
}

// NginxStopArg 强制停止nginx
type NginxStopArg struct {
	TaskId string
}

type NginxStopRse struct {
	TaskId string
	Rse    string
}

// NginxCheckConfArg 检查nginx配置
type NginxCheckConfArg struct {
	TaskId string
}

type NginxCheckConfRse struct {
	TaskId string
	Rse    string
}

// NginxConfUpdateArg  修改nginx配置文件
type NginxConfUpdateArg struct {
	TaskId string
	Path   string
	Data   string
}

type NginxConfUpdateRse struct {
	TaskId string
	Rse    string
}
