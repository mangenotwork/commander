package linux

import (
	"gitee.com/mangenotework/commander/common/cmd"
	"regexp"
	"strings"
)

// HaveDocker host 是否安装docker
// 执行 ps  -ef | grep docker 命令 如果返回有2条以上则安装
// ps  -aux | grep docker 也可以
// 返回第二个参数是pid 字符串类型
func HaveDocker() (bool, string) {
	rStr := cmd.LinuxSendCommand("ps -e|grep docker")
	if rStr == "" {
		return false, ""
	}
	//log.Println(rStr)
	rStrList := strings.Split(rStr, "\n")
	rList := make([]string, 0)
	for _, v := range rStrList {
		if v != "" {
			rList = append(rList, v)
		}
	}
	//log.Println(rList, len(rList))
	if len(rList) >= 1 {
		pid := ""
		pidList := strings.Split(rList[0], " ")
		for _, v := range pidList {
			if v != "" {
				pid = v
				break
			}
		}
		return true, pid
	}
	return false, ""
}

// CmdDockerVersion docker 版本  通过 docker version 命令获取
func CmdDockerVersion() string {
	isDocker, _ := HaveDocker()
	if !isDocker {
		return ""
	}
	rStr := cmd.LinuxSendCommand("docker version")
	if rStr != "" {
		reg := regexp.MustCompile(`Version:(.*?)\n`)
		sList := reg.FindAllString(rStr, -1)
		if len(sList) > 1 {
			return sList[0]
		}
	}
	return ""
}
