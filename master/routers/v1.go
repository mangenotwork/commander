package routers

import (
	"gitee.com/mangenotework/commander/master/handler"
)

func V1() {

	v1 := Router.Group("", HttpMiddleware())

	// ========================================================================================== pg (页面)
	v1.GET("/", handler.PGLogin)                                                  // 登录页面
	v1.POST("/register", handler.Register)                                        // 登录页面
	v1.POST("/login", handler.Login)                                              // 登录
	v1.GET("/home", AuthPG(), handler.PGIndex)                                    // 首页
	v1.GET("/docker/:slave", AuthPG(), handler.PGDockerManage)                    // salve 的docker 管理页面
	v1.GET("/slave/:slave", AuthPG(), handler.PGSlave)                            // salve 页面
	v1.GET("/container/log", AuthPG(), handler.PGContainerLog)                    // docker 容器的日志页面
	v1.GET("/container/monitor", AuthPG(), handler.PGContainerMonitor)            // docker 容器的实时监控性能页面 (需要引入图表)
	v1.GET("/slave/process/:slave", AuthPG(), handler.PGSlaveProcess)             // salve process 页面
	v1.GET("/slave/port/:slave", AuthPG(), handler.PGSlavePort)                   // salve port 页面
	v1.GET("/executable", AuthPG(), handler.PGExecutableManage)                   // 進程管理 页面
	v1.GET("/executable/log", AuthPG(), handler.PGExecutableLog)                  // 可執行文件的日志页面
	v1.GET("/cache", AuthPG(), handler.PGCache)                                   //緩存管理頁面
	v1.GET("/project", AuthPG(), handler.PGProjectManage)                         // 项目管理页面
	v1.GET("/project/container/:project", AuthPG(), handler.PGProjectContainer)   // 项目管理 - 项目的docker 容器页面
	v1.GET("/task", AuthPG(), handler.PGTaskPage)                                 // 任务页面
	v1.GET("/project/executable/:project", AuthPG(), handler.PGProjectExecutable) // 项目管理 - 项目的可执行任务任务
	v1.GET("/gateway", AuthPG(), handler.PGGateway)                               // 网关管理
	v1.GET("/monitor", AuthPG(), handler.PGMonitor)                               // 监控管理
	v1.GET("/monitor/:slave", AuthPG(), handler.MonitorSlave)                     // slave的实时监控页面
	v1.GET("/slave/env/:slave", AuthPG(), handler.PGSlaveEnv)                     // slave的环境变量页面
	v1.GET("/executable/dir", AuthPG(), handler.PGExecutableDir)                  // 查看可执行文件的目录结构
	v1.GET("/forward", AuthPG(), handler.PGForward)                               // 网络转发管理
	v1.GET("/nginx/:slave", AuthPG(), handler.PGNginx)                            // Nginx管理
	v1.GET("/deployed/:slave", AuthPG(), handler.PGEnvDeployed)                   // 环境部署
	v1.GET("/dir/:slave", AuthPG(), handler.PGDir)                                // TODO 目录&文件  功能: 文件上传 下载 删除 新建目录
	v1.GET("/ssh", AuthPG(), handler.PGSSH)                                       // ssh
	v1.GET("/clear", handler.ClearToken)                                          // 清除token
	v1.GET("/settings", handler.Settings)                                         // 设置页面

	// ========================================================================================== WebSocket
	Router.Any("/ws/notice", handler.WebSocketNotice)                // 通知
	Router.Any("/ws/container/log", handler.WebSocketContainerLog)   // docker 容器的日志
	Router.Any("/ws/executable/log", handler.WebSocketExecutableLog) // 可執行文件日誌
	Router.Any("/ws/ssh", handler.WebSocketTerminal)                 // 可執行文件日誌

	// ========================================================================================== api
	api := v1.Group("/api", AuthPG())

	// ========== slave host （主机相关信息与操作）
	{
		api.GET("/slave/list", handler.SlaveList)                 // slave 列表
		api.GET("/slave/info", handler.SlaveInfo)                 // slave 信息
		api.GET("/slave/process/list", handler.SlaveProcessList)  // 查看主机(slave)进程
		api.GET("/slave/env/list", handler.SlaveENVList)          // 查看主机(slave)环境变量
		api.GET("/slave/disk/info", handler.SlaveDiskInfo)        // 查看主机(slave)磁盘信息
		api.GET("/slave/path/info", handler.SlavePathInfo)        // 查看主机(slave)目录与文件
		api.GET("/slave/process/kill", handler.SlaveProcessKill)  // 结束一个进程	 [记录到操作]
		api.GET("/slave/port/info", handler.SlavePortInfo)        // 查看salve端口使用情況
		api.GET("/slave/process/info", handler.SlaveProcessInfo)  // 查看单个进程的详情
		api.GET("/slave/select", handler.PGSlaveSelect)           // salve 下拉框筛选条件
		api.GET("/slave/hosts", handler.SlaveHosts)               // 读取 hosts 文件
		api.POST("/slave/hosts/update", handler.SlaveHostsUpdate) // 修改 hosts 文件
		api.GET("/slave/dir", handler.SlaveDir)                   // 查看slave 目录与文件
		api.GET("/slave/cat", handler.SlaveCatFile)               // 查看slave文件内容
		api.GET("/slave/download", handler.SlaveFileDownload)     // 下载slave文件
		api.POST("/slave/upload", handler.SlaveFileUpload)        // 上传slave文件
		api.GET("/slave/mkdir", handler.SlaveMkdir)               // 创建目录
		api.GET("/slave/pack/dir", handler.SlavePackDir)          // 打包下载整个目录
		api.GET("/slave/decompress", handler.SlaveDecompress)     // 解压 slave 的压缩文件
		// TODO  查看進程使用的端口
		// TODO  查看slave运行时长
		// TODO  查看slave cpu 温度
		// TODO  增加slave环境变量	[记录到操作]
		// TODO  删除slave环境变量	[记录到操作]
		// TODO  修改slave环境变量	[记录到操作]
		// TODO  slave 是什么发行版本系统
	}

	// ========== task （任务）
	{
		api.GET("/task/list", handler.TaskList)     // 任务列表
		api.GET("/task/delete", handler.TaskDelete) // 任务删除
	}

	// ========== docker （docker相关操作接口）
	{
		api.GET("/docker/info", handler.DockerInfo)                   // docker 信息
		api.GET("/docker/have", handler.HaveDocker)                   // slave 是否有启动docker服务
		api.GET("/docker/ps", handler.DockerPS)                       // docker ps
		api.GET("/docker/images", handler.DockerImages)               // docker images
		api.POST("/docker/images/deploy", handler.DockerImagesDeploy) // docker images
		api.GET("/docker/stop", handler.DockerStop)                   // docker stop	[记录到任务] [记录到操作]
		api.GET("/docker/rm", handler.DockerRm)                       // docker rm			[记录到任务] [记录到操作]
		api.GET("/docker/rmi", handler.DockerRmi)                     // docker rmi		[记录到任务] [记录到操作]
		api.GET("/docker/pull", handler.DockerPull)                   // docker pull	[记录到任务] [记录到操作]
		api.GET("/docker/task/state", handler.GetTaskState)           // 查看task
		api.POST("/docker/run", handler.DockerRun)                    // dcoker run		[记录到任务] [记录到操作]
		api.GET("/docker/top", handler.DockerTop)                     // 查看容器进程  docker top [记录到操作]
		api.GET("/docker/rename", handler.DockerReName)               // 修改容器名称 docker rename 	[记录到任务] [记录到操作]
		api.GET("/docker/restart", handler.DockerRestart)             // 容器重启  docker restart		[记录到任务] [记录到操作]
		api.GET("/docker/pause", handler.DockerPause)                 // 容器暂停  docker pause		[记录到任务] [记录到操作]
		api.GET("/docker/states", handler.DockerStates)               // 返回所有容器cpu mem 信息
		api.GET("/docker/state", handler.DockerState)                 // 实时监控容器性能 [记录到操作]
		// TODO 查看容器的详情
		// TODO 容器网络列表，显示各个容器之间的网络配置。 [记录到操作]
	}

	// ========== executable   （可执行文件相关操作接口）
	{
		api.POST("/executable/upload", handler.ExecutableUpload)           //  上传可执行文件 [记录到操作]
		api.Any("/executable/download", handler.ExecutableDownload)        // 下载可执行文件 [记录到操作]
		Router.Any("/executable/download", handler.ExecutableDownload)     // 下载可执行文件 TODO slave 下载，需要安全
		api.GET("/executable/list", handler.ExecutableList)                // 查看可执行文件
		api.GET("/executable/delete", handler.ExecutableDelete)            // 删除可执行文件 [记录到操作]
		api.POST("/executable/deploy", handler.ExecutableDeploy)           // 部署可执行文件		[记录到任务] [记录到操作]
		api.GET("/executable/task", handler.ExecutableRunList)             // 已經執行的可執行文件
		api.GET("/executable/run/state", handler.ExecutableRunState)       // 已經執行的可執行文件的狀態
		api.GET("/executable/task/delete", handler.ExecutableTaskDelete)   // 刪除已經執行的可执行文件任務  如果正在執行無法刪除 [记录到操作]
		api.GET("/executable/task/run", handler.ExecutableTaskRun)         // 启动可执行文件任务		[记录到任务] [记录到操作]
		api.GET("/executable/task/kill", handler.ExecutableKill)           // 停止已經執行的可執行文件		[记录到任务] [记录到操作]
		api.GET("/executable/task/restart", handler.ExecutableTaskRestart) // 重啓已經執行的可执行文件進程 [记录到操作]
		api.GET("/executable/task/pid", handler.ExecutableTaskPIDInfo)     // 查看进程详情  没有执行则无法查看
		api.GET("/executable/dir", handler.ExecutableDir)                  // 查看可执行文件的目录结构
		api.GET("/executable/conf/file", handler.ExecutableConfFile)       //  查看可执行文件的目录结构并找出配置文件
		api.POST("/executable/conf/update", handler.ExecutableConfUpdate)  // 修改配置文件
		api.GET("/executable/task/log", handler.ExecutableTaskLog)         // TODO 查看可执行文件运行输出的终端打印日志
	}

	// ========== monitor （监控相关操作接口）
	{
		api.GET("/monitor/rule/list", handler.MonitorRuleList)      // 监控标准列表
		api.POST("/monitor/rule/create", handler.MonitorRuleCreate) // 新增监控标准 [记录到操作]
		api.GET("/monitor/alarm/list", handler.MonitorAlarmList)    // 查看报警列表
		api.GET("/monitor/alarm/del", handler.MonitorAlarmDel)      // 报警删除 [记录到操作]
		api.GET("/monitor/data", handler.MonitorData)               // 获取指定时间段性能数据
		// TODO 设置发送邮件通知 [记录到操作]
		// TODO 设置发送钉钉通知 [记录到操作]
	}

	// ========== project （项目相关操作接口）
	{
		// 项目管理 - executable项目
		api.POST("/project/executable/create", handler.ProjectExecutableCreate)    // 新建 executable项目	 [记录到操作]
		api.GET("/project/executable/run", handler.ProjectExecutableRun)           // TODO 部署executable项目	[记录到任务] [记录到操作]
		api.GET("/project/executable/list", handler.ProjectExecutableList)         // executable项目列表
		api.Any("/project/executable/download", handler.ProjectExecutableDownload) // 下载executable项目	[记录到操作]
		api.GET("/project/executable/task", handler.ProjectExecutableTaskList)     // executable项目执行任务列表
		// 项目管理 - docker容器部署项目
		api.POST("/project/docker/create", handler.ProjectDockerCreate)                    // 新建docker容器项目		[记录到任务] [记录到操作]
		api.POST("/project/docker/update", handler.ProjectDockerUpdate)                    // TODO 更新并重启容器项目	[记录到任务] [记录到操作]
		api.POST("/project/docker/update/image", handler.ProjectDockerUpdateImage)         // 更新项目的镜像，并重启		[记录到任务] [记录到操作]
		api.POST("/project/docker/update/duplicate", handler.ProjectDockerUpdateDuplicate) // 更新副本数量		[记录到任务] [记录到操作]
		api.GET("/project/docker/run", handler.ProjectDockerRun)                           // 部署docker项目		[记录到任务] [记录到操作]
		api.GET("/project/docker/container", handler.ProjectDockerContainer)               // 查看项目下的docker容器 [记录到操作]
		api.GET("/project/docker/list", handler.ProjectDockerList)                         // docker容器项目列表 [记录到操作]
		api.GET("/project/docker/container/ips", handler.ProjectDockerContainerIps)        // 获取容器的ip地址
		Router.GET("/project/docker/container/ips", handler.ProjectDockerContainerIps)     // 获取容器的ip地址
	}

	// ========== gateway （网关相关操作接口）
	// 网络服务层， 是一个网关，主要用于转发代理等
	{
		api.POST("/gateway/new", handler.GatewayNew)                // TODO 新建网关
		api.GET("/gateway/run", handler.GatewayRun)                 // 启动一个网关		[记录到操作]
		api.GET("/gateway/list", handler.GatewayList)               // 网关列表
		api.GET("/gateway/delete", handler.GatewayDelete)           // 删除一个网关 [记录到操作]
		api.POST("/gateway/update/port", handler.GatewayUpdatePort) // 修改网关端口映射	[记录到操作]
	}

	// ========== cache （緩存相关操作接口）
	{
		api.GET("/cache/size", handler.CacheSize)
		api.GET("/cache/list", handler.CacheList)
		api.GET("/cache/delete", handler.CacheDelete) // 删除缓存 [记录到操作]
		// TODO 定时任务，定时清理新能采集数据，避免持续增大， 可配置限制大小
	}

	// ========== operatenotes （operate notes; 记录操作的接口）
	{
		api.GET("/operate/list", handler.OperateList)
		api.GET("/operate/delete", handler.OperateDelete)
	}

	// ========== deployed 部署软件相关
	{
		api.GET("/deployed/install/docker", handler.DeployedInstallDocker) // 安装docker
		api.GET("/deployed/remove/docker", handler.DeployedRemoveDocker)   // 卸载docker
		api.GET("/deployed/install/nginx", handler.DeployedInstallNginx)   // 安装nginx
		api.GET("/deployed/remove/nginx", handler.DeployedRemoveNginx)     // 卸载nginx
	}

	// ========== proxy&forward  网络代理与网络转发服务
	{
		api.POST("/proxy/http/create", handler.ProxyHttpCreate)          // 创建http/s代理
		api.GET("/proxy/http/list", handler.ProxyHttpList)               // 获取http/s代理列表
		api.GET("/proxy/http/stop", handler.ProxyHttpStop)               // 获取http/s代理暂停
		api.GET("/proxy/http/continue", handler.ProxyHttpContinue)       // 获取http/s代理继续
		api.GET("/proxy/http/remove", handler.ProxyHttpRemove)           // 删除http/s代理
		api.POST("/proxy/socket5/create", handler.ProxySocket5Create)    // 创建socket5代理
		api.GET("/proxy/socket5/list", handler.ProxySocket5List)         // 获取socket5代理列表
		api.GET("/proxy/socket5/stop", handler.ProxySocket5Stop)         // 获取http/s代理暂停
		api.GET("/proxy/socket5/continue", handler.ProxySocket5Continue) // 获取http/s代理继续
		api.GET("/proxy/socket5/remove", handler.ProxySocket5Remove)     // 删除socket5代理
		api.POST("/forward/tcp/create", handler.ForwardTCPCreate)        // 创建TCP转发
		api.GET("/forward/tcp/list", handler.ForwardTCPList)             // 获取TCP转发列表
		api.GET("/forward/tcp/remove", handler.ForwardTCPRemove)         // 删除TCP转发
		api.GET("/forward/tcp/switch", handler.ForwardTCPSwitch)         // 切换TCP转发目标
		api.GET("/forward/tcp/stop", handler.ForwardTCPCutoff)           // 切断TCP转发
		api.GET("/forward/tcp/continue", handler.ForwardTCPRenew)        // 恢复TCP转发
		api.POST("/forward/udp/create", handler.ForwardUDPCreate)        // TODO 创建UDP转发
		api.GET("/forward/udp/list", handler.ForwardUDPList)             // TODO 获取UDP转发列表
		api.GET("/forward/udp/remove", handler.ForwardUDPRemove)         // TODO 删除UDP转发
		api.GET("/forward/udp/switch", handler.ForwardUDPSwitch)         // TODO 切换UDP转发目标
		api.GET("/forward/udp/cutoff", handler.ForwardUDPCutoff)         // TODO 切断UDP转发
		api.GET("/forward/udp/renew", handler.ForwardUDPRenew)           // TODO 恢复UDP转发
		api.POST("/proxy/ssh/create", handler.ProxySSHCreate)            // TODO 创建ssh代理
		api.GET("/proxy/ssh/list", handler.ProxySSHList)                 // TODO 获取ssh代理列表
		api.GET("/proxy/ssh/remove", handler.ProxySSHRemove)             // TODO 删除ssh代理
	}

	// ========== nginx
	{
		api.GET("/nginx/info", handler.NginxInfo)               // 获取nginx信息
		api.GET("/nginx/start", handler.NginxStart)             // 启动nginx
		api.GET("/nginx/reload", handler.NginxReload)           // 重启nginx -s reload
		api.GET("/nginx/quit", handler.NginxQuit)               // 停止nginx nginx -s quit
		api.GET("/nginx/stop", handler.NginxStop)               // 强制停止nginx nginx -s stop
		api.GET("/nginx/check", handler.NginxCheckConf)         // 检查nginx配置
		api.POST("/nginx/conf/update", handler.NginxConfUpdate) // 修改nginx配置文件
	}

	// ========== auth  （身份验证相关接口）
	{
		// TODO ....
	}

	// ========== NATPenetration   （NAT penetration; 内网穿透）
	{
		// TODO ....
	}

	// ========== CDN servers
	{
		// TODO ....
	}

	// ========== terminal
	{
		// TODO linux  ssh 远程终端    可参考   GoTTY web终端
	}

	// =========== test （用于测试或实验的接口）
	Router.GET("/test/notice", handler.TestNotice)
	Router.GET("/test/docker/state", handler.TestDockerState)

}
