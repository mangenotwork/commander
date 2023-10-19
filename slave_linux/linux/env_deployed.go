package linux

import (
	"gitee.com/mangenotework/commander/common/cmd"
	"gitee.com/mangenotework/commander/common/logger"
	"log"
	"os/exec"
	"strings"
)

func IsYum() bool {
	cmd := exec.Command("yum")
	err := cmd.Run()
	if err != nil {
		if strings.Index(err.Error(), "not found") == -1 {
			log.Println("有该命令")
			return true
		}
		log.Println("cmd.Run() failed with %s\n", err)
	}
	log.Println("没有该命令")
	return false
}

func IsApt() bool {
	cmd := exec.Command("apt")
	err := cmd.Run()
	if err != nil {
		if strings.Index(err.Error(), "not found") == -1 {
			log.Println("有该命令")
			return true
		}
		log.Println("cmd.Run() failed with %s\n", err)
	}
	log.Println("没有该命令")
	return false
}

// DeployedDockerCE 安装部署 docker ce  卸载旧docker
func DeployedDockerCE() string {
	rse := ""

	if IsApt() {
		logger.Info("执行 yum 安装 docker ")
		// sudo apt-get remove docker.io docker-engine
		stp1 := cmd.LinuxSendCommand("sudo apt-get remove -y docker.io docker-engine")
		logger.Info(stp1)
		rse += stp1

		// sudo apt-get update # 更新源
		stp2 := cmd.LinuxSendCommand("sudo apt-get update -y")
		logger.Info(stp2)
		rse += stp2

		// sudo apt-get install docker-ce docker-ce-cli containerd.io # 安装
		stp3 := cmd.LinuxSendCommand("sudo apt-get install -y docker-ce docker-ce-cli containerd.io")
		logger.Info(stp3)
		rse += stp3

		// docker version
		stp4 := cmd.LinuxSendCommand("docker version")
		logger.Info(stp4)
		rse += stp4
	}

	if IsYum() {
		logger.Info("执行 apt 安装 docker ")
		// yum -y remove docker docker-common docker-selinux docker-engine
		stp1 := cmd.LinuxSendCommand("yum -y remove docker docker-common docker-selinux docker-engine")
		logger.Info(stp1)
		rse += stp1

		// yum install -y yum-utils device-mapper-persistent-data lvm2 # 配置yum 源码
		stp2 := cmd.LinuxSendCommand("yum install -y yum-utils device-mapper-persistent-data lvm2")
		logger.Info(stp2)
		rse += stp2

		// yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
		stp3 := cmd.LinuxSendCommand("yum install -y yum-utils device-mapper-persistent-data lvm2")
		logger.Info(stp3)
		rse += stp3

		// yum install -y docker-ce                # 默认安装最新版本 docker
		stp4 := cmd.LinuxSendCommand("yum install -y docker-ce")
		logger.Info(stp4)
		rse += stp4

		// systemctl start docker                  # 启动docker
		stp5 := cmd.LinuxSendCommand("systemctl start docker")
		logger.Info(stp5)
		rse += stp5

		// systemctl enable docker         # 开机启动
		stp6 := cmd.LinuxSendCommand("systemctl enable docker")
		logger.Info(stp6)
		rse += stp6

		// docker version                          # 查看docker版本号
		stp7 := cmd.LinuxSendCommand("docker version")
		logger.Info(stp7)
		rse += stp7
	}

	return rse
}

// RemoveDocker 卸载 Docker
func RemoveDocker() string {
	rse := ""

	if IsApt() {
		// apt-get remove -y docker-ce
		stp1 := cmd.LinuxSendCommand("apt-get remove -y docker-ce")
		log.Println(stp1)
		rse += stp1

		// apt-get remove -y docker-ce-cli
		stp2 := cmd.LinuxSendCommand("apt-get remove -y docker-ce-cli")
		log.Println(stp2)
		rse += stp2
	}

	if IsYum() {
		// yum remove -y docker-ce-cli
		stp1 := cmd.LinuxSendCommand("yum remove -y docker-ce-cli")
		log.Println(stp1)
		rse += stp1

		// yum remove -y docker-scan-plugin
		stp2 := cmd.LinuxSendCommand("yum remove -y docker-scan-plugin")
		log.Println(stp2)
		rse += stp2
	}

	return rse
}
