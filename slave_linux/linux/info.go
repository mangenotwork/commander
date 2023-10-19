package linux

import (
	"fmt"
	"gitee.com/mangenotework/commander/common/cmd"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
	"os"
	"regexp"
	"runtime"
	"strings"
)

// GetHostName 获取host 命名
func GetHostName() string {
	name, err := os.Hostname()
	if err != nil {
		name = "null"
	}
	return name
}

// GetSysType 获取host 系统类型
func GetSysType() string {
	return runtime.GOOS
}

// GetSysArch 获取系统架构
func GetSysArch() string {
	return runtime.GOARCH
}

// GetCpuCoreNumber 获取cpu核心数
func GetCpuCoreNumber() string {
	return fmt.Sprintf("%d核", runtime.GOMAXPROCS(0))
}

// ProcVersion 从 /proc/version 当前系统运行的内核版本号
func ProcVersion() (string, string) {
	rStr := cmd.LinuxSendCommand("cat /proc/version")
	if rStr == "" {
		return rStr, ""
	}
	version := ""
	reg := regexp.MustCompile(`Linux version(.*?)-`)
	sList := reg.FindStringSubmatch(rStr)
	if len(sList) > 1 {
		version = sList[1]
	}
	return rStr, utils.DeletePreAndSufSpace(version)
}

func GetCPUName() string {
	cpuName := ""
	cpuInfoList := ProcCPUinfo()
	if len(cpuInfoList) >= 2 {
		cpuName = cpuInfoList[1]["model name\t"]
	}
	return cpuName
}

// ProcCPUinfo 从 /proc/cpuinfo 获取cpu相关信息
func ProcCPUinfo() (cpuinfos []map[string]string) {
	cpuinfos = make([]map[string]string, 0)
	rStr := cmd.LinuxSendCommand("cat /proc/cpuinfo")
	if rStr == "" {
		return
	}
	rStrList := strings.Split(rStr, "processor")
	for _, v := range rStrList {
		v = "processor" + v
		vList := strings.Split(v, "\n")
		data := make(map[string]string, 0)
		for _, i := range vList {
			d := strings.Split(i, ":")
			if len(d) == 2 {
				key := utils.DeletePreAndSufSpace(d[0])
				vlue := utils.DeletePreAndSufSpace(d[1])
				data[key] = vlue
			}
		}
		cpuinfos = append(cpuinfos, data)
	}
	return
}

func GetMEMTotal() string {
	memInfo := ProcMeminfo()
	return fmt.Sprintf("%d MB", memInfo.MemTotal/1024)
}

// ProcMeminfo 从 /proc/meminfo 中读取内存
func ProcMeminfo() (mem *entity.ProcMemInfo) {
	rStr := cmd.LinuxSendCommand("cat /proc/meminfo")
	if rStr == "" {
		return
	}
	data := make(map[string]string, 0)
	rStrList := strings.Split(rStr, "\n")
	for _, v := range rStrList {
		d := strings.Split(v, ":")
		if len(d) == 2 {
			key := utils.DeletePreAndSufSpace(d[0])
			vlue := utils.DeletePreAndSufSpace(d[1])
			data[key] = vlue
		}
	}

	memTotal := utils.Str2Int64(data["MemTotal"])
	memFree := utils.Str2Int64(data["MemFree"])
	buffers := utils.Str2Int64(data["Buffers"])
	cached := utils.Str2Int64(data["Cached"])
	mem = &entity.ProcMemInfo{
		MemTotal:   memTotal,
		MemUsed:    memTotal - memFree - buffers - cached,
		MemFree:    memFree + buffers + cached,
		MemBuffers: buffers,
		MemCached:  cached,
	}
	return
}

// GetSystemDF 通过df 采样磁盘的基本
func GetSystemDF() (diskinfos []*entity.DiskInfo, totalStr string) {
	var allTotal int64 = 0
	diskinfos = make([]*entity.DiskInfo, 0)
	rStr := cmd.LinuxSendCommand("df -m")
	if rStr == "" {
		return
	}
	rStrList := strings.Split(rStr, "\n")
	if len(rStrList) < 2 {
		return
	}
	for _, v := range rStrList[1:len(rStrList)] {
		if v == "" {
			continue
		}
		//log.Println(v)
		vList := strings.Split(v, " ")
		nList := []string{}
		for _, n := range vList {
			if n == "" {
				continue
			}
			nList = append(nList, n)
		}
		//log.Println(nList, len(nList))
		if len(nList) > 5 {
			diskinfo := &entity.DiskInfo{
				DiskName:    nList[0],
				DistType:    "",
				DistTotalMB: nList[1],
			}
			//log.Println(nList[1])
			total := utils.Num2Int(nList[1])
			allTotal = allTotal + int64(total)
			diskinfo.DistUse = &entity.DiskUseInfo{
				Total: total,
				Free:  utils.Num2Int(nList[3]),
				Rate:  float32(utils.Str2Int64(nList[4])),
			}
			diskinfos = append(diskinfos, diskinfo)
		}
	}
	totalStr = utils.Int642Str(allTotal)
	//log.Println(rStr)
	return
}

func GetPortInfo() []*entity.PortInfo {
	rse := cmd.LinuxSendCommand("netstat -aptn")
	logger.Info(rse)
	row := strings.Split(rse, "\n")
	logger.Info(row)

	portList := make([]*entity.PortInfo, 0)

	for _, v := range row {
		logger.Info(v)
		vList := strings.Split(v, " ")

		nList := make([]string, 0)
		for _, n := range vList {
			if n != "" {
				nList = append(nList, n)
			}
		}
		logger.Info(nList, len(nList))

		if len(nList) != 7 {
			continue
		}
		pid := "-"
		pname := "-"
		if nList[6] != "-" {
			pList := strings.Split(nList[6], "/")
			pid = pList[0]
			pname = pList[1]
		}

		portList = append(portList, &entity.PortInfo{
			ProtoType:      nList[0],
			LocalAddress:   nList[3],
			ForeignAddress: nList[4],
			State:          nList[5],
			PID:            pid,
			PName:          pname,
		})

		logger.Info("______________________")
	}

	return portList
}

// TODO
// linux 查看 CPU 的速度（兆赫兹）  lscpu | grep -i mhz
// linux CPU 的详细信息   lscpu 或 lshw -C cpu
