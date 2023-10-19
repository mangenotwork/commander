package handler

import (
	"net"
	"time"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/slve_linux/linux"
)

// Health 心跳
// TODO 上报主机实时采集的信息
func Health(conn *net.UDPConn) {
	for {
		performance := GatherPerformance()
		buf, err := protocol.DataEncoder(performance)
		if err != nil {
			logger.Error(err)
			continue
		}
		packate, err := protocol.Packet(protocol.CMD_Health, "", buf)
		if err != nil {
			logger.Error(err)
			continue
		}
		_, _ = conn.Write(packate)
		//log.Println("发送心跳")
		time.Sleep(5 * time.Second)
	}
}

func GatherPerformance() *entity.SlavePerformance {
	performance := &entity.SlavePerformance{}

	//采集以下数据
	// CPU使用率
	performance.CPU, performance.CPUCore = linux.CPURate(1 * time.Second)
	//logger.Info("CPU使用率 = ", performance.CPU, performance.CPUCore )

	//	内存使用率
	performance.MEM = linux.ProcMeminfo()
	//logger.Info("内存使用率 = ", performance.MEM)

	//  磁盘使用率
	performance.Disk, _ = linux.GetSystemDF()
	//logger.Info("磁盘使用率 = ", performance.Disk)

	//	网络IO
	//	Tx float64
	//	Rx float64
	performance.NetWork = linux.ProcNetDev(1 * time.Second)
	//logger.Info("网络IO = ", performance.NetWork)

	//  连接数
	//	ConnectNum float64
	performance.ConnectNum = linux.GetTcpConnCount()
	//logger.Info("连接数 = ", performance.ConnectNum)

	return performance
}
