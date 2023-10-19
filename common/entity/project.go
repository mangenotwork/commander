package entity

// ProjectDocker docker容器项目
type ProjectDocker struct {
	Name         string `json:"docker_name"`       // 项目名称
	Note         string `json:"docker_note"`       // 项目备注
	IsGateway    string `json:"docker_is_gateway"` // 是否启动网关 0不启动， 1启动
	Port         string `json:"docker_port"`       // 占用端口
	Image        string `json:"docker_image"`      // 镜像
	User         string `json:"docker_user"`       // 账号
	Password     string `json:"docker_password"`   // 密码
	Env          string `json:"docker_env"`        // 环境变量
	Volume       string `json:"docker_volume"`     // Volume
	Duplicate    string `json:"docker_duplicate"`  // 副本数量
	CreateTime   string `json:"create_time"`
	GatewaySlave string `json:"gateway_slave"`
	GatewayPort  string `json:"gateway_port"`
	UpdateTime   string `json:"update_time"`
}

// ProjectExecutable 物理机可执行文件项目
type ProjectExecutable struct {
	Name        string `json:"executable_name"`      // 项目名称
	Note        string `json:"executable_note"`      // 项目备注
	Port        string `json:"executable_port"`      // 占用端口
	Cmd         string `json:"executable_cmd"`       // 执行命令
	Env         string `json:"executable_env"`       // 环境变量
	Duplicate   string `json:"executable_duplicate"` // 副本数量
	Sys         string `json:"executable_sys"`       // 运行平台
	Path        string `json:"executable_dir"`       // 可执行文件项目的目录
	ZipFilePath string `json:"executable_zip"`       // 可执行文件项目的压缩文件
	CreateTime  string `json:"create_time"`
}

// ProjectExecutableRunArg 部署并运行物理机可执行文件项目参数
type ProjectExecutableRunArg struct {
	ProjectName string
	Slave       string
	Port        string
	Env         string
	TaskId      string
	Cmd         string
}

// ProjectExecutableRunRse 部署并运行物理机可执行文件项目结果返回
type ProjectExecutableRunRse struct {
	ProjectName string
	Slave       string
	TaskId      string
	Pid         string
	Cmd         string
}

// RegisterIpToGatewayArg 注册地址到网关
type RegisterIpToGatewayArg struct {
	Key string
	Ip  string
}

// RegisterIpUpdateArg 通知网关更新
type RegisterIpUpdateArg struct {
	Project    string
	TargetPort string
}
