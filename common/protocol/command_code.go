package protocol

type CommandCode uint8

const (
	CMD_Health         CommandCode = 0x01 // salve 心跳
	CMD_ReportHostInfo CommandCode = 0x02 // salve 上报主机信息
	CMD_Reply          CommandCode = 0x03 // salve 应答

	// TODO 预留  0x04 ~ 0x19

	CMD_HaveDocker           CommandCode = 0x20 // 是否有docker
	CMD_DockerInfo           CommandCode = 0x21 // docker info
	CMD_DockerPS             CommandCode = 0x22 // docker ps
	CMD_DockerImages         CommandCode = 0x23 // docker images
	CMD_DockerPull           CommandCode = 0x24 // docker pull
	CMD_DockerRun            CommandCode = 0x25 // docker run
	CMD_DockerStop           CommandCode = 0x26 // docker stop
	CMD_DockerRm             CommandCode = 0x27 // docker rm
	CMD_DockerRmi            CommandCode = 0x28 // docker rmi
	CMD_DockerRename         CommandCode = 0x29 // docker rename
	CMD_DockerStateS         CommandCode = 0x30 //docker state  多个容器, 单个传用一个即可
	CMD_ContainerTop         CommandCode = 0x31 // 查看容器进程
	CMD_ContainerLog         CommandCode = 0x32 // 查看容器日志
	CMD_ContainerRename      CommandCode = 0x33 // 修改容器名称
	CMD_ContainerRestart     CommandCode = 0x34 // 容器重启
	CMD_ContainerPause       CommandCode = 0x35 // 容器暂停
	CMD_SlaveProcessList     CommandCode = 0x36 // 获取slave进程列表
	CMD_SlaveENVList         CommandCode = 0x37 // 获取slave环境变量
	CMD_SlaveDiskInfo        CommandCode = 0x38 // 获取slave磁盘信息
	CMD_SlavePathInfo        CommandCode = 0x39 // 获取slave目录与文件
	CMD_ExecutableDeploy     CommandCode = 0x40 // 通知 slave下载可执行文件並且部署運行
	CMD_SlavePortInfo        CommandCode = 0x41 // 查看salve端口使用情況
	CMD_ProcessKill          CommandCode = 0x42 // 關閉指定進程
	CMD_ExecutableRunState   CommandCode = 0x43 // 查看可执行文件的運行狀態
	CMD_ExecutablePIDLog     CommandCode = 0x44 // 查看可执行文件日誌 - 实时查看
	CMD_ExecutableKill       CommandCode = 0x45 // 结束可执行文件的进程
	CMD_SlaveProcessInfo     CommandCode = 0x46 // TODO 查看进程信息
	CMD_ExecutablePIDLogFile CommandCode = 0x47 // TODO 查看可执行文件日誌 的所有日志文件列表
	CMD_ProjectExecutableRun CommandCode = 0x48 // TODO 部署并运行可执行文件项目
	CMD_GatewayRun           CommandCode = 0x49 // 部署并运行网关服务
	CMD_RegisterIpToGateway  CommandCode = 0x50 // 将容器地址注册到网关上
	CMD_DelRegisterIPGateway CommandCode = 0x51 // 将指定容器地址删除注册到网关上的地址
	CMD_GatewayDel           CommandCode = 0x52 // 删除一个网关
	CMD_RegisterIPUpdate     CommandCode = 0x53 // 通知网关更新
	CMD_DockerImageRun       CommandCode = 0x54 // 运行指定镜像
	CMD_SlaveHosts           CommandCode = 0x55 // 读取hosts文件
	CMD_SlaveHostsUpdate     CommandCode = 0x56 // 修改hosts文件
	CMD_WarningNotice        CommandCode = 0x57 // 发送一个警告通知
	CMD_EnvDeployedCheck     CommandCode = 0x58 // 判断是否安装这些软件
	CMD_InstallDocker        CommandCode = 0x59 // 安装docker
	CMD_RemoveDocker         CommandCode = 0x60 // 卸载docker
	CMD_InstallNginx         CommandCode = 0x61 // 安装nginx
	CMD_RemoveNginx          CommandCode = 0x62 // 卸载nginx
	CMD_CreateHttpProxy      CommandCode = 0x63 // 创建并部署http/s代理
	CMD_CreateSocket5Proxy   CommandCode = 0x64 // 创建并部署socket代理
	CMD_CreateTCPForward     CommandCode = 0x65 // 创建并部署TCP转发
	CMD_UpdateHttpProxy      CommandCode = 0x66 // 修改http/s代理
	CMD_UpdateSocket5Proxy   CommandCode = 0x67 // 修改socket5代理
	CMD_UpdateTCPForward     CommandCode = 0x68 // 修改tcp转发
	CMD_CreateUDPForward     CommandCode = 0x69 // TODO 创建UDP转发
	CMD_UpdateUDPForward     CommandCode = 0x70 // TODO 修改UDP转发
	CMD_CreateSSHProxy       CommandCode = 0x71 // TODO 创建并部署ssh代理
	CMD_UpdateSSHProxy       CommandCode = 0x72 // TODO 修改ssh代理
	CMD_GetSlavePathInfo     CommandCode = 0x73 // 获取slave指定路径的目录结构与文件
	CMD_SlaveCatFile         CommandCode = 0x74 // 查看slave指定文件内容
	CMD_SlaveMkdir           CommandCode = 0x75 // 在slave创建目录
	CMD_SlaveDecompress      CommandCode = 0x76 // 解压slave 的压缩文件
	CMD_NginxInfo            CommandCode = 0x77 // 获取nginx信息
	CMD_NginxStart           CommandCode = 0x78 // 启动nginx
	CMD_NginxReload          CommandCode = 0x79 // 重启nginx
	CMD_NginxQuit            CommandCode = 0x80 // 停止nginx
	CMD_NginxStop            CommandCode = 0x81 // 强制停止nginx
	CMD_NginxCheckConf       CommandCode = 0x82 // 检查nginx 配置
	CMD_NginxConfUpdate      CommandCode = 0x83 // 更新nginx 配置文件
)

func (c CommandCode) Chinese() string {
	return CommandCodeChineseMap[c]
}

var CommandCodeChineseMap map[CommandCode]string = map[CommandCode]string{
	CMD_Health:               "CMD_Health: 心跳包",
	CMD_ReportHostInfo:       "CMD_ReportHostInfo: salve 上报主机信息",
	CMD_Reply:                "CMD_Reply: salve 应答",
	CMD_HaveDocker:           "CMD_HaveDocker: 是否有docker",
	CMD_DockerInfo:           "CMD_DockerInfo: docker info",
	CMD_DockerPS:             "CMD_DockerPS: docker ps",
	CMD_DockerImages:         "CMD_DockerImages: docker images",
	CMD_DockerPull:           "CMD_DockerPull: docker pull",
	CMD_DockerRun:            "CMD_DockerRun: docker run",
	CMD_DockerStop:           "CMD_DockerStop: docker stop",
	CMD_DockerRm:             "CMD_DockerRm: docker rm",
	CMD_DockerRmi:            "CMD_DockerRmi: docker rmi",
	CMD_DockerRename:         "CMD_DockerRename: docker rename",
	CMD_DockerStateS:         "CMD_DockerStateS: docker state  多个容器, 单个传用一个即可",
	CMD_ContainerTop:         "CMD_ContainerTop: 查看容器进程",
	CMD_ContainerLog:         "CMD_ContainerLog: 查看容器日志",
	CMD_ContainerRename:      "CMD_ContainerRename: 修改容器名称",
	CMD_ContainerRestart:     "CMD_ContainerRestart: 容器重启",
	CMD_ContainerPause:       "CMD_ContainerPause: 容器暂停",
	CMD_SlaveProcessList:     "CMD_SlaveProcessList: 获取slave进程列表",
	CMD_SlaveENVList:         "CMD_SlaveENVList: 获取slave环境变量",
	CMD_SlaveDiskInfo:        "CMD_SlaveDiskInfo: 获取slave磁盘信息",
	CMD_SlavePathInfo:        "CMD_SlavePathInfo: 获取slave目录与文件",
	CMD_ExecutableDeploy:     "CMD_ExecutableDeploy: 通知 slave下载可执行文件並且部署運行",
	CMD_SlavePortInfo:        "CMD_SlavePortInfo: 查看salve端口使用情況",
	CMD_ProcessKill:          "CMD_ProcessKill: 關閉指定進程",
	CMD_ExecutableRunState:   "CMD_ExecutableRunState: 查看可执行文件的運行狀態",
	CMD_ExecutablePIDLog:     "CMD_ExecutablePIDLog: 查看可执行文件日誌 - 实时查看",
	CMD_ExecutableKill:       "CMD_ExecutableKill: 结束可执行文件的进程",
	CMD_SlaveProcessInfo:     "CMD_SlaveProcessInfo: 查看进程信息",
	CMD_ExecutablePIDLogFile: "CMD_ExecutablePIDLogFile: 查看可执行文件日誌 的所有日志文件列表",
	CMD_ProjectExecutableRun: "CMD_ProjectExecutableRun: 部署并运行可执行文件项目",
	CMD_GatewayRun:           "CMD_GatewayRun: 部署并运行网关服务",
	CMD_RegisterIpToGateway:  "CMD_RegisterIpToGateway: 将容器地址注册到网关上",
	CMD_DelRegisterIPGateway: "CMD_DelRegisterIPGateway: 将指定容器地址删除注册到网关上的地址",
	CMD_GatewayDel:           "CMD_GatewayDel: 删除一个网关",
	CMD_RegisterIPUpdate:     "CMD_RegisterIPUpdate: 通知网关更新",
	CMD_EnvDeployedCheck:     "CMD_EnvDeployedCheck: 判断是否安装这些软件",
	CMD_InstallDocker:        "CMD_InstallDocker: 安装docker",
	CMD_RemoveDocker:         "CMD_RemoveDocker: 卸载docker",
	CMD_InstallNginx:         "CMD_InstallNginx: 安装nginx",
	CMD_RemoveNginx:          "CMD_RemoveNginx: 卸载nginx",
	CMD_CreateHttpProxy:      "CMD_CreateHttpProxy: 创建并部署http/s代理",
	CMD_CreateSocket5Proxy:   "CMD_CreateSocket5Proxy: 创建并部署socket代理",
	CMD_CreateTCPForward:     "CMD_CreateTCPForward:  创建并部署TCP转发",
	CMD_UpdateHttpProxy:      "CMD_UpdateHttpProxy: 修改http/s代理",
	CMD_UpdateSocket5Proxy:   "CMD_UpdateSocket5Proxy: 修改socket5代理",
	CMD_UpdateTCPForward:     "CMD_UpdateTCPForward: 修改tcp转发",
	CMD_CreateUDPForward:     "CMD_CreateUDPForward: 创建UDP转发",
	CMD_UpdateUDPForward:     "CMD_UpdateUDPForward: 修改UDP转发",
	CMD_CreateSSHProxy:       "CMD_CreateSSHProxy: 创建并部署ssh代理",
	CMD_UpdateSSHProxy:       "CMD_UpdateSSHProxy: TODO 修改ssh代理",
	CMD_GetSlavePathInfo:     "CMD_GetSlavePathInfo: 获取slave指定路径的目录结构与文件",
	CMD_SlaveCatFile:         "CMD_SlaveCatFile: 查看slave指定文件内容",
	CMD_SlaveMkdir:           "CMD_SlaveMkdir :在slave创建目录",
	CMD_SlaveDecompress:      "CMD_SlaveDecompress: 解压slave 的压缩文件",
	CMD_NginxInfo:            "CMD_NginxInfo: 获取nginx信息",
	CMD_NginxStart:           "CMD_NginxStart: 启动nginx",
	CMD_NginxReload:          "CMD_NginxReload: 重启nginx",
	CMD_NginxQuit:            "CMD_NginxQuit: 停止nginx",
	CMD_NginxStop:            "CMD_NginxStop: 强制停止nginx",
	CMD_NginxCheckConf:       "CMD_NginxCheckConf: 检查nginx 配置",
	CMD_NginxConfUpdate:      "CMD_NginxConfUpdate: 更新nginx 配置文件",
}
