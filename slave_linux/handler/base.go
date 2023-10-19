package handler

import (
	"net"

	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
)

var InitFunc = []func(conn *net.UDPConn){Health, ReportHostInfo}

var ErrFunc = []func(){CloseAllContainerLogIO}

func InitHandler() protocol.Handler {
	return protocol.Handler{
		protocol.CMD_Health:               Hello,
		protocol.CMD_ReportHostInfo:       HostInfo,
		protocol.CMD_HaveDocker:           HaveDocker,
		protocol.CMD_DockerInfo:           DockerInfo,
		protocol.CMD_DockerPS:             DockerPs,
		protocol.CMD_DockerImages:         DockerImages,
		protocol.CMD_DockerPull:           DockerPull,
		protocol.CMD_DockerRun:            DockerRun,
		protocol.CMD_DockerStop:           DockerStop,
		protocol.CMD_DockerRm:             DockerRm,
		protocol.CMD_DockerRmi:            DockerRmi,
		protocol.CMD_ContainerLog:         ContainerLog,
		protocol.CMD_ContainerTop:         ContainerTop,
		protocol.CMD_ContainerRename:      ContainerRename,
		protocol.CMD_ContainerRestart:     ContainerRestart,
		protocol.CMD_ContainerPause:       ContainerPause,
		protocol.CMD_DockerStateS:         DockerStateS,
		protocol.CMD_SlaveProcessList:     SlaveProcessList,
		protocol.CMD_SlaveENVList:         SlaveENVList,
		protocol.CMD_SlaveDiskInfo:        SlaveDiskInfo,
		protocol.CMD_SlavePathInfo:        SlavePathInfo,
		protocol.CMD_ExecutableDeploy:     ExecutableDeploy,
		protocol.CMD_SlavePortInfo:        SlavePortInfo,
		protocol.CMD_ProcessKill:          ProcessKill,
		protocol.CMD_ExecutableRunState:   ExecutableRunState,
		protocol.CMD_ExecutablePIDLog:     ExecutablePIDLog,
		protocol.CMD_ExecutableKill:       ExecutableKill,
		protocol.CMD_SlaveProcessInfo:     SlaveProcessInfo,
		protocol.CMD_ProjectExecutableRun: ProjectExecutableRun,
		protocol.CMD_GatewayRun:           GatewayRun,
		protocol.CMD_RegisterIpToGateway:  RegisterIpToGateway,
		protocol.CMD_DelRegisterIPGateway: DelRegisterIPGateway,
		protocol.CMD_GatewayDel:           GatewayDel,
		protocol.CMD_RegisterIPUpdate:     RegisterIPUpdate,
		protocol.CMD_DockerImageRun:       DockerImageRun,
		protocol.CMD_SlaveHosts:           SlaveHosts,
		protocol.CMD_SlaveHostsUpdate:     SlaveHostsUpdate,
		protocol.CMD_EnvDeployedCheck:     EnvDeployedCheck,
		protocol.CMD_InstallDocker:        InstallDocker,
		protocol.CMD_RemoveDocker:         RemoveDocker,
		protocol.CMD_InstallNginx:         InstallNginx,
		protocol.CMD_RemoveNginx:          RemoveNginx,
		protocol.CMD_CreateHttpProxy:      CreateHttpProxy,
		protocol.CMD_UpdateHttpProxy:      UpdateHttpProxy,
		protocol.CMD_CreateSocket5Proxy:   CreateSocket5Proxy,
		protocol.CMD_UpdateSocket5Proxy:   UpdateSocket5Proxy,
		protocol.CMD_CreateTCPForward:     CreateTCPForward,
		protocol.CMD_UpdateTCPForward:     UpdateTCPForward,
		protocol.CMD_GetSlavePathInfo:     GetSlavePathInfo,
		protocol.CMD_SlaveCatFile:         SlaveCatFile,
		protocol.CMD_SlaveMkdir:           SlaveMkdir,
		protocol.CMD_SlaveDecompress:      SlaveDecompress,
		protocol.CMD_NginxInfo:            NginxInfo,
		protocol.CMD_NginxStart:           NginxStart,
		protocol.CMD_NginxReload:          NginxReload,
		protocol.CMD_NginxQuit:            NginxQuit,
		protocol.CMD_NginxStop:            NginxStop,
		protocol.CMD_NginxCheckConf:       NginxCheckConf,
		protocol.CMD_NginxConfUpdate:      NginxConfUpdate,
	}
}

// Hello 测试通讯应答
func Hello(ctx *protocol.HandlerCtx) {
	data := ctx.Stream.Data
	logger.Info("Heool,Heool，Heool，")
	logger.Info(string(data))

	// 应答
	err := ctx.Send([]byte("收到"))
	if err != nil {
		logger.Info(err)
		return
	}
}
