package linux

import (
	"fmt"
	"gitee.com/mangenotework/commander/common/cmd"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// HasNginx 是否安装nginx
// nginx -v是向stderr写入了数据，所以从stdout是拿不到数据的
func HasNginx() (bool, string) {
	exeCmd := exec.Command("nginx", "-v")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	if strings.Index(rse, "nginx version") == -1 {
		return false, "未安装Nginx"
	}
	return true, rse
}

// AptInstallNginx apt 安装nginx
func AptInstallNginx() string {
	rse := ""
	if IsApt() {
		// 更新源
		stp1 := cmd.LinuxSendCommand("sudo apt-get update -y")
		log.Println(stp1)
		rse += stp1

		// 安装nginx
		stp2 := cmd.LinuxSendCommand("sudo apt-get install nginx -y")
		log.Println(stp2)
		rse += stp2

		// 查看安装
		stp3 := cmd.LinuxSendCommand("nginx -v")
		log.Println(stp3)
		rse += stp3
	}
	return rse
}

// YumInstallNginx yum 安装nginx
func YumInstallNginx() string {
	rse := ""
	if IsYum() {
		// 下载 epel-release
		stp1 := cmd.LinuxSendCommand("sudo yum -y install epel-release")
		log.Println(stp1)
		rse += stp1

		// 更新源
		stp2 := cmd.LinuxSendCommand("sudo yum -y update")
		log.Println(stp2)
		rse += stp2

		// 安装
		stp3 := cmd.LinuxSendCommand("sudo yum -y install nginx")
		log.Println(stp3)
		rse += stp3

		// 查看安装
		stp4 := cmd.LinuxSendCommand("nginx -v")
		log.Println(stp4)
		rse += stp4
	}
	return rse
}

// RemoveNginx 卸载nginx
func RemoveNginx() string {
	rse := ""
	if IsApt() {
		// sudo apt-get --purge -y remove nginx
		stp1 := cmd.LinuxSendCommand("sudo apt-get --purge -y remove nginx")
		log.Println(stp1)
		rse += stp1

		// sudo apt-get --purge -y remove nginx-common
		stp2 := cmd.LinuxSendCommand("sudo apt-get --purge -y remove nginx-common")
		log.Println(stp2)
		rse += stp2

		// sudo apt-get --purge -y remove nginx-core
		stp3 := cmd.LinuxSendCommand("sudo apt-get --purge -y remove nginx-core")
		log.Println(stp3)
		rse += stp3
	}

	if IsYum() {
		// yum remove -y nginx
		stp1 := cmd.LinuxSendCommand("yum remove -y nginx")
		log.Println(stp1)
		rse += stp1
	}
	return rse
}

func DeployedNginx() string {
	rse := AptInstallNginx()
	rse = YumInstallNginx()
	return rse
}

// NginxConfPath nginx -t  获取nginx配置文件路径
func NginxConfPath() string {
	exeCmd := exec.Command("nginx", "-t")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	for _, s := range strings.Split(rse, " ") {
		if strings.Index(s, ".conf") != -1 {
			return s
		}
	}
	return ""
}

// NginxPath whereis nginx  获取nginx路径
func NginxPath() string {
	exeCmd := exec.Command("whereis", "nginx")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxPid ps -C nginx -o pid
func NginxPid() []string {
	exeCmd := exec.Command("ps", "-C", "nginx", "-o", "pid")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	pids := make([]string, 0)
	rse = strings.Replace(rse, "\t", "", -1)
	rse = strings.Replace(rse, " ", "", -1)
	for _, s := range strings.Split(rse, "\n") {
		if len(s) < 1 {
			continue
		}
		if s == "PID" {
			continue
		}
		pids = append(pids, s)
	}
	return pids
}

// NginxServiceStatus service nginx status
func NginxServiceStatus() string {
	exeCmd := exec.Command("service", "nginx", "status")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxReload nginx -s reload 重启
func NginxReload() string {
	exeCmd := exec.Command("nginx", "-s", "reload")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxStart nginx 启动
func NginxStart() string {
	exeCmd := exec.Command("nginx")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxStop nginx -s stop 强制停止
func NginxStop() string {
	exeCmd := exec.Command("nginx", "-s", "stop")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxQuit nginx -s quit 停止
func NginxQuit() string {
	exeCmd := exec.Command("nginx", "-s", "quit")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxServiceStart service nginx start #启动
func NginxServiceStart() string {
	exeCmd := exec.Command("service", "nginx", "start")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxServiceStop service nginx stop #停止
func NginxServiceStop() string {
	exeCmd := exec.Command("service", "nginx", "stop")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxServiceRestart service nginx restart #重启
func NginxServiceRestart() string {
	exeCmd := exec.Command("service", "nginx", "restart")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	rse := string(out)
	return rse
}

// NginxT nginx -t 测试配置文件
func NginxT() string {
	exeCmd := exec.Command("nginx", "-t")
	exeCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exeCmd.CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	return string(out)
}
