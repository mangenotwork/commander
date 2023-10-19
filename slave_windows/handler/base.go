package handler

import (
	"fmt"
	"log"
	"net"
	"time"

	"gitee.com/mangenotework/commander/common/protocol"
)

var InitFunc = []func(conn *net.UDPConn){Health, ReportHostInfo}

var ErrFunc = []func(){}

func InitHandler() protocol.Handler {
	return protocol.Handler {
		protocol.CMD_Health: Hello,
		protocol.CMD_ReportHostInfo: CMDReportHostInfo,
		protocol.CMD_HaveDocker: HaveDocker,
		protocol.CMD_DockerInfo: DockerInfo,
		protocol.CMD_DockerPS: DockerPs,
		protocol.CMD_DockerImages: DockerImages,
		protocol.CMD_DockerPull: DockerPull,
		protocol.CMD_DockerRun: DockerRun,
		protocol.CMD_DockerStop: DockerStop,
		protocol.CMD_DockerRm: DockerRm,
		protocol.CMD_DockerRmi: DockerRmi,
		protocol.CMD_ContainerLog: ContainerLog,
		protocol.CMD_ContainerTop: ContainerTop,
		protocol.CMD_ContainerRename: ContainerRename,
		protocol.CMD_ContainerRestart: ContainerRestart,
		protocol.CMD_ContainerPause: ContainerPause,
		protocol.CMD_DockerStateS: DockerStateS,
		protocol.CMD_SlaveProcessList: SlaveProcessList,
		protocol.CMD_SlaveENVList: SlaveENVList,
		protocol.CMD_SlaveDiskInfo: SlaveDiskInfo,
		//protocol.CMD_SlavePathInfo: SlavePathInfo,
		//protocol.CMD_ExecutableDeploy: ExecutableDeploy,
		//protocol.CMD_SlavePortInfo: SlavePortInfo,
		//protocol.CMD_ProcessKill: ProcessKill,
		//protocol.CMD_ExecutableRunState: ExecutableRunState,
		//protocol.CMD_ExecutablePIDLog: ExecutablePIDLog,
		//protocol.CMD_ExecutableKill: ExecutableKill,
		//protocol.CMD_SlaveProcessInfo: SlaveProcessInfo,
	}
}

// Health 心跳
// TODO 上报主机实时采集的信息
func Health(conn *net.UDPConn) {
	for {

		//performance := GatherPerformance()
		//buf, err := protocol.DataEncoder(performance)
		//if  err != nil {
		//	fmt.Println(err)
		//}

		buf := []byte("心跳")
		packate, err := protocol.Packet(protocol.CMD_Health, "", buf)
		if err != nil {
			log.Println(err)
			continue
		}
		_,_=conn.Write(packate)
		log.Println("发送心跳")
		time.Sleep(5*time.Second)
	}
}

// ReportHostInfo 上报主机信息
func ReportHostInfo(conn *net.UDPConn) {
	info := hostInfo()
	buf, err := protocol.GobEncoder(info)
	if  err != nil {
		fmt.Println(err)
		return
	}
	packate, err := protocol.Packet(protocol.CMD_ReportHostInfo, "", buf)
	if err != nil {
		log.Println(err)
	}
	_,_=conn.Write(packate)
}

func Hello(ctx *protocol.HandlerCtx) {
	data := ctx.Stream.Data
	log.Println("Heool,Heool，Heool，")
	log.Println(string(data))

	// 应答
	err := ctx.Send([]byte("收到"))
	if err != nil {
		log.Println(err)
		return
	}
}