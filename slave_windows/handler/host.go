package handler

import (
	"fmt"
	"log"
	"os"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/slave_windows/win"
)

// hostInfo 获取主机信息
func hostInfo() *entity.HostInfo {
	info := &entity.HostInfo{}
	//主机名称
	info.HostName = win.GetComputerName()
	log.Println("主机名称 = ", info.HostName)
	//系统平台
	info.SysType = win.WindowsGetOsInfo()
	log.Println("系统平台 = ", info.SysType)
	//系统版本 os_name+版号
	info.OsName = win.WindowsGetOsInfo()
	log.Println("系统版本 os_name+版号 = ", info.OsName)
	//TODO 系统架构
	info.SysArchitecture = utils.GetSysArch()
	log.Println("系统架构 = ", info.SysArchitecture)
	//CPU核心数
	info.CpuCoreNumber = utils.GetCpuCoreNumber()
	log.Println("CPU核心数 = ", info.CpuCoreNumber)
	//CPU name
	info.CpuName = win.GetCpuName()
	//info.CpuName = ""
	log.Println("CPU name = ", info.CpuName)
	//CPU ID
	info.CpuID = win.GetCpuId()
	log.Println("CPU ID = ", info.CpuID)
	//主板ID
	info.BaseBoardID = win.GetBaseBoardID()
	log.Println("主板ID = ", info.BaseBoardID)
	//内存总大小 MB
	info.MemTotal = win.WindowsGetMemoryTotal()
	log.Println("内存总大小 MB = ", info.MemTotal)
	//磁盘信息, //磁盘总大小 MB
	diskTotal := 0
	diskInfo := win.WindowsGetDiskInfo()
	for _, v := range diskInfo {
		diskTotal += v.DistUse.Total
	}
	info.Disk = diskInfo
	info.DiskTotal = fmt.Sprintf("%dMB", diskTotal)
	log.Println("磁盘信息, //磁盘总大小 MB = ", info.Disk, info.DiskTotal)
	return info
}

func CMDReportHostInfo(ctx *protocol.HandlerCtx){
	info := hostInfo()
	buf, err := protocol.GobEncoder(info)
	if  err != nil {
		fmt.Println(err)
		return
	}
	err = ctx.Send(buf)
	if err != nil {
		log.Println("err = ", err)
	}
}


// SlaveProcessList 获取进程列表
func SlaveProcessList(ctx *protocol.HandlerCtx){
	psList := win.GetProcessList(utils.Any2Int(string(ctx.Stream.Data)))
	log.Println("psList = ", psList)
	buf, err := protocol.DataEncoder(psList)
	if  err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if  err != nil {
		fmt.Println(err)
	}
}

// SlaveENVList 获取环境变量
func SlaveENVList(ctx *protocol.HandlerCtx) {
	envs := os.Environ()
	buf, err := protocol.DataEncoder(envs)
	if  err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if  err != nil {
		fmt.Println(err)
	}
}

// SlaveDiskInfo 磁盘信息
func SlaveDiskInfo(ctx *protocol.HandlerCtx) {
	data := win.WindowsGetDiskInfo()
	buf, err := protocol.DataEncoder(data)
	if  err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if  err != nil {
		fmt.Println(err)
	}
}