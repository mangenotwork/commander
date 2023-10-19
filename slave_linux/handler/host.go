package handler

import (
	"bufio"
	"context"
	"fmt"
	"gitee.com/mangenotework/commander/slve_linux/dao"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	cmd2 "gitee.com/mangenotework/commander/common/cmd"
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/enum"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/slve_linux/gateway"
	"gitee.com/mangenotework/commander/slve_linux/linux"
)

// ReportHostInfo 上报主机信息
func ReportHostInfo(conn *net.UDPConn) {
	info := hostInfo()
	buf, err := protocol.GobEncoder(info)
	if err != nil {
		fmt.Println(err)
		return
	}
	packate, err := protocol.Packet(protocol.CMD_ReportHostInfo, "", buf)
	if err != nil {
		logger.Error(err)
	}
	_, _ = conn.Write(packate)
}

func hostInfo() *entity.HostInfo {
	info := &entity.HostInfo{
		SlaveVersion: enum.SlaveVersion,
	}
	//主机名称
	info.HostName = linux.GetHostName()
	logger.Info("主机名称 = ", info.HostName)
	//系统平台
	info.SysType = linux.GetSysType()
	logger.Info("系统平台 = ", info.SysType)
	//系统版本 os_name+版号
	info.OsName, info.OsNum = linux.ProcVersion()
	logger.Info("系统版本 os_name+版号 = ", info.OsName)
	//系统架构
	info.SysArchitecture = linux.GetSysArch()
	logger.Info("系统架构 = ", info.SysArchitecture)
	//CPU核心数
	info.CpuCoreNumber = linux.GetCpuCoreNumber()
	logger.Info("CPU核心数 = ", info.CpuCoreNumber)
	//CPU name
	info.CpuName = linux.GetCPUName()
	logger.Info("CPU name = ", info.CpuName)
	//CPU ID
	info.CpuID = linux.GetCPUIDFromLinux()
	logger.Info("CPU ID = ", info.CpuID)
	//主板ID
	info.BaseBoardID = "TODO..."
	logger.Info("主板ID = ", info.BaseBoardID)
	//内存总大小 MB
	info.MemTotal = linux.GetMEMTotal()
	logger.Info("内存总大小 MB = ", info.MemTotal)
	//磁盘信息, //磁盘总大小 MB
	info.Disk, info.DiskTotal = linux.GetSystemDF()
	logger.Info("磁盘信息, //磁盘总大小 MB = ", info.Disk, info.DiskTotal)
	// docker 相关
	ok, pid := linux.HaveDocker()
	if ok {
		info.HasDocker = "true pid:" + pid
		info.DockerVersion = linux.CmdDockerVersion()
	} else {
		info.HasDocker = "false"
		info.DockerVersion = ""
	}
	info.FileServerPort = conf.SlaveConf.FileServer.Port
	info.FileServerSecret = conf.SlaveConf.FileServer.Secret
	//系统运行时间
	info.RunTime, info.LdleTime = linux.ProcUptime()
	return info
}

// HostInfo 上报主机信息
func HostInfo(ctx *protocol.HandlerCtx) {
	info := hostInfo()

	buf, err := protocol.GobEncoder(info)
	if err != nil {
		fmt.Println(err)
		return
	}

	//logger.Info("ctx.Cmd = ", ctx.Cmd)

	if ctx.Cmd == 0 {

	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error("err = ", err)
	}
}

// HaveDocker 是否安装了docker
func HaveDocker(ctx *protocol.HandlerCtx) {
	info := linux.CmdDockerVersion()
	if info == "" {
		info = "没有安装docker"
	}
	logger.Info("是否安装docker = ", info)
	err := ctx.Send([]byte(info))
	if err != nil {
		logger.Error("err = ", err)
		return
	}
}

// SlaveProcessList 查看进程列表
func SlaveProcessList(ctx *protocol.HandlerCtx) {

	psList := linux.GetProcessList(utils.Any2Int(string(ctx.Stream.Data)))
	logger.Info("psList = ", psList)
	buf, err := protocol.DataEncoder(psList)
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}
}

// SlaveENVList 系统环境变量列表
func SlaveENVList(ctx *protocol.HandlerCtx) {
	envs := os.Environ()
	// TODO 命令行获取

	buf, err := protocol.DataEncoder(envs)
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}
}

// SlaveDiskInfo 获取磁盘信息
func SlaveDiskInfo(ctx *protocol.HandlerCtx) {
	data, _ := linux.GetSystemDF()
	buf, err := protocol.DataEncoder(data)
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}
}

// SlavePathInfo 获取路径下的目录与文件信息
func SlavePathInfo(ctx *protocol.HandlerCtx) {
	data := make([]*entity.FileInfo, 0)
	pwd := string(ctx.Stream.Data)
	fileInfoList, err := ioutil.ReadDir(pwd)
	logger.Info("fileInfoList = ", fileInfoList, err)
	for _, v := range fileInfoList {
		logger.Info(v)
		fInfo := &entity.FileInfo{
			Name:    v.Name(),
			Size:    v.Size(),
			Mode:    int(v.Mode()),
			ModTime: v.ModTime().String(),
			IsDir:   v.IsDir(),
			Sys:     v.Sys(),
		}
		logger.Info(fInfo)
		data = append(data, fInfo)
	}

	buf, err := protocol.DataEncoder(data)
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}
}

func LinuxSendCommand(command string) (opStr string) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", command)
	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		logger.Error("ERR stdout : ", stdoutErr)
	}
	defer cancel()
	defer stdout.Close()
	if startErr := cmd.Start(); startErr != nil {
		logger.Error("ERR Start : ", startErr)
	}
	opBytes, opBytesErr := ioutil.ReadAll(stdout)
	if opBytesErr != nil {
		//log.Println(string(opBytes))
		opStr = ""
	}
	opStr = string(opBytes)
	//log.Println(opStr)
	cmd.Wait()
	return
}

// ExecutableDeploy 下载可执行文件
// TODO 捕獲stdout， 用於後面查看日誌
func ExecutableDeploy(ctx *protocol.HandlerCtx) {
	pull := entity.ExecutableDeployArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &pull)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	logger.Info("pull = ", pull)

	// 拉取项目
	url := conf.SlaveConf.MasterHttp + "/executable/download?name=" + pull.DownloadFile
	logger.Info("url,path = ", url, conf.SlaveConf.ExeStoreHousePath)
	exePath, err := utils.HTTPDownLoad(url, conf.SlaveConf.ExeStoreHousePath)
	if err != nil {
		//TODO 发送错误
		logger.Error(err)
	}
	// 解压
	exeDir := strings.Split(exePath, ".")[0] + "/"

	err = utils.DeCompressZIP(exePath, exeDir)
	if err != nil {
		logger.Error("解压失败 : ", err)
		return
	}

	// 执行程序
	logger.Info("exePath = ", exePath)
	logger.Info("exePath dir = ", exeDir)
	logger.Info("cmd  = ", pull.Cmd)
	cmdCtx, cancel := context.WithCancel(context.Background())
	cmd2.LinuxSendCommand("sudo chmod 777 " + exeDir + " -R ")
	cmder := exec.CommandContext(cmdCtx, "/bin/bash", "-c", pull.Cmd)
	cmder.Env = strings.Split(pull.Env, ";")
	cmder.Dir = exeDir

	defer cancel()

	stdout, stdoutErr := cmder.StdoutPipe()
	if stdoutErr != nil {
		logger.Error("ERR stdout : ", stdoutErr)
	}

	// 日誌讀取 方案一
	//ExecutableIOReadCloserStore(pull.TaskId, &StdoutObj{
	//	Stdout: stdout,
	//	Stop: make(chan int),
	//})

	startErr := cmder.Start()
	if startErr != nil {
		logger.Error("ERR Start : ", startErr)
	}
	logger.Info("PID = ", cmder.Process.Pid)
	//cmd.ProcessState.ExitCode()

	results := &entity.ExecutableDeployTask{
		Slave:          pull.Slave,
		TaskId:         pull.TaskId,
		Command:        pull.Cmd,
		Env:            pull.Env,
		ExecutableName: pull.DownloadFile,
		Arg:            pull.Arg,
		Time:           utils.NowTimeStr(),
		PID:            cmder.Process.Pid,
		State:          "請刷新...",
		Note:           pull.Note,
	}

	// 1.本地保存這個執行的事件
	_ = new(dao.DaoSlaveExecutableTask).Set([]byte(pull.TaskId), results)

	// 2.執行後保存這個進程(後期對這個進程的操作)
	ExecutableDeployProcessStore(pull.TaskId, cmder)

	// 3.返回給 master 信息
	go func() {
		buf, err := protocol.DataEncoder(results)
		if err != nil {
			fmt.Println(err)
		}
		err = ctx.Send(buf)
		if err != nil {
			fmt.Println(err)
		}
	}()

	var f *os.File
	// 按当天进行保存日志
	filename := conf.SlaveConf.ExeStoreHouseLogs + pull.TaskId + "_" + utils.BeginDayUnixStr() + ".log"
	ok, _ := utils.PathExists(filename)
	if ok { //如果文件存在
		f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0666) //打开文件
		logger.Info("文件存在")
	} else {
		f, err = os.Create(filename) //创建文件
		logger.Info("文件不存在")
	}
	if err != nil {
		logger.Error("寫入日誌失敗 err = ", err)
		panic("寫入日誌失敗 err = " + err.Error())
	}

	// 日誌讀取 方案二 收集輸出
	readout := bufio.NewReader(stdout)
	go func() {

		outputBytes := make([]byte, 200)
		for {
			n, err := readout.Read(outputBytes) //获取屏幕的实时输出(并不是按照回车分割，所以要结合sumOutput)
			if err != nil {
				if err == io.EOF {
					break
				}
				logger.Error(err)
				return
			}
			output := string(outputBytes[:n])
			//logger.Info(output) //寫入文件

			if n > 0 {
				//logger.Info("写入文件 ： ", filename)
				_, writeErr := io.WriteString(f, output) //写入文件(字符串)
				if writeErr != nil {
					logger.Error("write err = ", writeErr)
				}
			}

			//time.Sleep(1*time.Second)
		}
	}()

	// TODO : 启动一个监视器，监视这个进程的状态，并通知给master
	go func() {
		pid := utils.StringValue(cmder.Process.Pid)
		for {
			logger.Info("pid = ", pid)
			isHave := linux.ProcessIsHave(pid)
			if isHave {
				err = ctx.SendCmd(protocol.CMD_ExecutableRunState, []byte(enum.ExecutableStateExecuting))
			} else {
				err = ctx.SendCmd(protocol.CMD_ExecutableRunState, []byte(enum.ExecutableStateDiscontinued))
				notice := fmt.Sprintf("项目: %s | pid: %s --> 停止运行", pull.DownloadFile, pid)
				logger.Error(notice)
				// TODO 发送通知
				_ = ctx.SendCmd(protocol.CMD_WarningNotice, []byte(notice))
				break
			}
			if err != nil {
				logger.Error(err)
			}
			time.Sleep(3 * time.Second)
		}
	}()

	// 等待：等待命令退出
	err = cmder.Wait()
	if err != nil {
		logger.Info("结束！！！")
		logger.Error("pid err = ", err)
		err = f.Close()
		if err != nil {
			logger.Error(err)
		}

		// TODO
		// 這裡可以收到進程中斷信息   pid err =  signal: killed

		// 3.守護這個進程(如果是slave掉了，重啓後需要恢復這些進程); 如果設置恢復（進程掉了後，自動恢復）
	}

}

// 保存可執行文件的進程
var executableDeployProcess sync.Map

// ExecutableDeployProcessStore taskId 任務id
func ExecutableDeployProcessStore(taskId string, process *exec.Cmd) {
	executableDeployProcess.Store(taskId, process)
}

func ExecutableDeployProcessLoad(taskId string) *exec.Cmd {
	val, ok := executableDeployProcess.Load(taskId)
	if ok {
		return val.(*exec.Cmd)
	}
	return nil
}

func ExecutableDeployProcessDelete(taskId string) {
	executableDeployProcess.Delete(taskId)
}

func SlavePortInfo(ctx *protocol.HandlerCtx) {
	logger.Info("SlavePortInfo ......")
	results := linux.GetPortInfo()
	buf, err := protocol.DataEncoder(results)
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}
}

func SlaveProcessInfo(ctx *protocol.HandlerCtx) {
	pid := string(ctx.Stream.Data)

	info := &entity.ProcessInfo{}

	//某一进程Cpu使用率的计算
	pcpu := linux.ProcessProcStat(pid)
	logger.Info("_________________________\n 某一进程Cpu使用率的计算: ", pcpu)
	info.PCPU = pcpu

	//获取启动当前进程的完整命令
	cmdStr := linux.ProcPIDCmdline(pid)
	logger.Info("_________________________\n 获取启动当前进程的完整命令: ", cmdStr)
	info.Cmd = cmdStr

	//当前进程的环境变量列表
	environList := linux.ProcPIDEnviron(pid)
	logger.Info("_________________________\n 当前进程的环境变量列表:", environList)
	info.Environ = environList

	////当前进程所使用的每一个受限资源的软限制硬限制和管理单元
	//log.Println("_________________________\n 当前进程所使用的每一个受限资源的软限制硬限制和管理单元:")
	//linux.ProcPIDLimits(pid)

	//当前进程关联到的每个可执行文件和库文件在内存中的映射区域及其访问权限所组成的列表
	//log.Println("_________________________\n 当前进程关联到的每个可执行文件和库文件在内存中的映射区域及其访问权限所组成的列表:")
	//linux.ProcPIDMaps(pid)

	//当前进程的状态信息，包含一系统格式化后的数据列
	statusStr := linux.ProcPIDStatus(pid)
	logger.Info("_________________________\n 当前进程的状态信息，包含一系统格式化后的数据列: ", statusStr)
	info.StatusTxt = statusStr

	buf, err := protocol.DataEncoder(info)
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}

}

// ===================================================  進程讀取日誌 方案一
// 實時讀取，保存 ReadCloser

var executableIOReadCloser sync.Map

type StdoutObj struct {
	Stdout io.ReadCloser
	Stop   chan int
}

// ExecutableIOReadCloserStore taskId 任務id
func ExecutableIOReadCloserStore(taskId string, stdout *StdoutObj) {
	executableIOReadCloser.Store(taskId, stdout)
}

func ExecutableIOReadCloserLoad(taskId string) *StdoutObj {
	val, ok := executableIOReadCloser.Load(taskId)
	if ok {
		return val.(*StdoutObj)
	}
	return nil
}

func ExecutableIOReadCloserDelete(taskId string) {
	executableIOReadCloser.Delete(taskId)
}

type ExecutablePIDLogArg struct {
	TaskId string
	Cmd    string
}

// ExecutablePIDLogBUG BUG: 沒有解決關閉goroutine
func ExecutablePIDLogBUG(ctx *protocol.HandlerCtx) {
	arg := &ExecutablePIDLogArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}

	stdout := ExecutableIOReadCloserLoad(arg.TaskId)

	if arg.Cmd != "close" {
		readout := bufio.NewReader(stdout.Stdout)
		go func() {
			outputBytes := make([]byte, 200)
			for {
				select {
				case <-stdout.Stop:
					logger.Info("logIO.Stop ...... ")
					return
				default:
					n, err := readout.Read(outputBytes) //获取屏幕的实时输出(并不是按照回车分割，所以要结合sumOutput)
					if err != nil {
						if err == io.EOF {
							break
						}
						logger.Error(err)
					}
					output := string(outputBytes[:n])
					logger.Info(output) //输出到master
					time.Sleep(1 * time.Second)
				}
			}
		}()
	} else {
		logger.Info("ExecutablePIDLog  Close ...... ")

		stdout.Stop <- 1
	}

}

// ExecutablePIDLog1 bug: 開多個窗口，沒有同時顯示
func ExecutablePIDLog1(ctx *protocol.HandlerCtx) {
	taskId := string(ctx.Stream.Data)
	stdout := ExecutableIOReadCloserLoad(taskId)
	readout := bufio.NewReader(stdout.Stdout)
	go func() {
		outputBytes := make([]byte, 200)
		for {
			n, err := readout.Read(outputBytes) //获取屏幕的实时输出(并不是按照回车分割，所以要结合sumOutput)
			if err != nil {
				if err == io.EOF {
					break
				}
				logger.Error(err)
				return
			}
			output := outputBytes[:n]
			logger.Info(output) //输出到master
			if n > 0 {
				logger.Info(string(output))
				_ = ctx.Send(output)
			}
			//time.Sleep(1*time.Second)
		}

	}()
}

// ===================================================  進程讀取日誌 方案二

// ExecutablePIDLog 讀取日誌文件
func ExecutablePIDLog(ctx *protocol.HandlerCtx) {
	taskId := string(ctx.Stream.Data)
	// 按当天进行实时读取日志
	filename := conf.SlaveConf.ExeStoreHouseLogs + taskId + "_" + utils.BeginDayUnixStr() + ".log"
	file, err := os.Open(filename)
	if err != nil {
		logger.Error(err)
		return
	}
	defer func() {
		err = file.Close() //close after checking err
	}()
	fileStat, err := file.Stat()
	if err != nil {
		logger.Error(err)
		return
	}
	fileSize := fileStat.Size() //72849354767
	//log.Println(fileSize)
	offset := fileSize - 1024*10
	if fileSize < 1024*10 {
		offset = 0
	}
	lastLine := make([]byte, 1024*10)
	_, err = file.ReadAt(lastLine, offset)
	if err != nil {
		logger.Error(err)
	}
	//logger.Info("lastLine = ", lastLine)
	_ = ctx.Send(lastLine)
}

// ProcessKill 關閉進程
func ProcessKill(ctx *protocol.HandlerCtx) {
	arg := entity.ProcessKillArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	logger.Info("arg = ", arg)
	if arg.Value == "" {
		arg.Value = "9"
	}
	rse := linux.Kill(arg.PID, "-"+arg.Value)
	err = ctx.Send([]byte(rse))
	if err != nil {
		fmt.Println(err)
	}
}

// ExecutableKill 关闭可执行文件
func ExecutableKill(ctx *protocol.HandlerCtx) {
	arg := entity.ExecutableKillArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	logger.Info("arg = ", arg)
	if arg.Value == "" {
		arg.Value = "9"
	}
	rse := linux.Kill(arg.PID, "-"+arg.Value)

	data := &entity.ExecutableKillRse{
		Rse:    rse,
		TaskId: arg.TaskId,
	}

	buf, err := protocol.DataEncoder(data)
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}
}

func ExecutableRunState(ctx *protocol.HandlerCtx) {

	arg := &entity.ExecutableRunStateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("ExecutableRunState err  = ", err)
	}

	pid := arg.PId
	logger.Info("pid = ", pid)
	isHave := linux.ProcessIsHave(pid)

	rse := &entity.ExecutableRunStateRse{
		PId:    arg.PId,
		TaskId: arg.TaskId,
	}

	if isHave {
		rse.Rse = enum.ExecutableStateExecuting
	} else {
		rse.Rse = enum.ExecutableStateDiscontinued
	}

	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}

}

func ProjectExecutableRun(ctx *protocol.HandlerCtx) {
	arg := &entity.ProjectExecutableRunArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	logger.Info("arg = ", arg)
	// 1. 拉取项目压缩文件，并解压
	url := conf.SlaveConf.MasterHttp + "/project/executable/download?project=" + arg.ProjectName
	logger.Info("url,path = ", url, conf.SlaveConf.ProjectExeStoreHousePath)
	exePath, err := utils.HTTPDownLoad(url, conf.SlaveConf.ProjectExeStoreHousePath)
	if err != nil {
		//TODO 发送错误
		logger.Error(err)
		return
	}
	dst := path.Join(conf.SlaveConf.ProjectExeStoreHousePath, arg.ProjectName)
	err = utils.DecompressionZipFile(exePath, dst)
	if err != nil {
		//TODO 发送错误
		logger.Error(err)
		return
	}

	// TODO 端口守护

	// 2. 执行命令
	// 3. 返回数据， 包含可执行文件数据 pid等等
	// 执行程序
	logger.Info("dst = ", dst)
	cmd2.LinuxSendCommand("chmod 777 " + dst + " -R")

	logger.Info("arg.Cmd = ", arg.Cmd)
	cmdStr := dst + "/" + arg.Cmd
	logger.Info("cmdStr = ", cmdStr)
	cmdCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(cmdCtx, "/bin/bash", "-c", cmdStr)
	defer cancel()

	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		logger.Error("ERR stdout : ", stdoutErr)
	}

	startErr := cmd.Start()
	if startErr != nil {
		logger.Error("ERR Start : ", startErr)
	}
	logger.Info("PID = ", cmd.Process.Pid)

	// 2.執行後保存這個進程(後期對這個進程的操作)
	ExecutableDeployProcessStore(arg.TaskId, cmd)

	// 3.返回給 master 信息
	go func() {

		results := entity.ProjectExecutableRunRse{
			ProjectName: arg.ProjectName,
			Slave:       arg.Slave,
			TaskId:      arg.TaskId,
			Pid:         utils.StringValue(cmd.Process.Pid),
			Cmd:         arg.Cmd,
		}

		buf, err := protocol.DataEncoder(results)
		if err != nil {
			fmt.Println(err)
		}
		err = ctx.Send(buf)
		if err != nil {
			fmt.Println(err)
		}
	}()

	var f *os.File
	// 按当天进行保存日志
	filename := conf.SlaveConf.ExeStoreHouseLogs + arg.TaskId + "_" + utils.BeginDayUnixStr() + ".log"
	ok, _ := utils.PathExists(filename)
	if ok { //如果文件存在
		f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0666) //打开文件
		fmt.Println("文件存在")
	} else {
		f, err = os.Create(filename) //创建文件
		fmt.Println("文件不存在")
	}
	if err != nil {
		logger.Error("寫入日誌失敗 err = ", err)
		panic("寫入日誌失敗 err = " + err.Error())
	}

	// 日誌讀取 方案二 收集輸出
	readout := bufio.NewReader(stdout)
	go func() {

		outputBytes := make([]byte, 200)
		for {
			n, err := readout.Read(outputBytes) //获取屏幕的实时输出(并不是按照回车分割，所以要结合sumOutput)
			if err != nil {
				if err == io.EOF {
					break
				}
				logger.Error(err)
				return
			}
			output := string(outputBytes[:n])
			logger.Info(output) //寫入文件

			if n > 0 {
				logger.Info("写入文件 ： ", filename)
				_, writeErr := io.WriteString(f, output) //写入文件(字符串)
				if writeErr != nil {
					logger.Error("write err = ", writeErr)
				}
			}

			//time.Sleep(1*time.Second)
		}
	}()

	// 等待：等待命令退出
	err = cmd.Wait()
	if err != nil {
		logger.Error("pid err = ", err)
		err = f.Close()
		if err != nil {
			logger.Error(err)
		}

		// TODO
		// 這裡可以收到進程中斷信息   pid err =  signal: killed

		// 3.守護這個進程(如果是slave掉了，重啓後需要恢復這些進程); 如果設置恢復（進程掉了後，自動恢復）
	}

}

// GatewayRun 部署并运行网关服务
// 缺点: 强依赖 master, master相等于发现服务
func GatewayRun(ctx *protocol.HandlerCtx) {
	arg := &entity.GatewayArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	// 记录网关数据，持久化记录网关服务， 当主机重启后自动启动网关
	gatewayBase := &entity.GatewayBase{
		Ports:       arg.Ports,
		ProjectName: arg.ProjectName,
		LVS:         arg.LVS,
		LVSModel:    arg.LVSModel,
		IsClose:     "0",
		Create:      utils.NowTimeStr(),
	}
	logger.Info("持久化网关数据..... ", arg.ProjectName, gatewayBase)
	err = new(dao.DaoGateway).Set([]byte(arg.ProjectName), gatewayBase)
	if err != nil {
		logger.Error("持久化数据失败 : ", err)
	}
	// 启动网关
	gateway.RunGateway(arg)
	buf, err := protocol.DataEncoder("网关已经启动")
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}
}

// RegisterIpToGateway 将容器网络地址注册到网关上
func RegisterIpToGateway(ctx *protocol.HandlerCtx) {
	arg := &entity.RegisterIpToGatewayArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	logger.Info("将容器网络地址注册到网关上 = ", arg.Key, arg.Ip)
	gateway.SetIps(arg.Key, arg.Ip)
}

// DelRegisterIPGateway 删除指定注册到网关上的地址
func DelRegisterIPGateway(ctx *protocol.HandlerCtx) {
	arg := &entity.RegisterIpToGatewayArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	logger.Info("删除指定注册到网关上的地址 = ", arg.Key, arg.Ip)
	gateway.DelIp(arg.Key, arg.Ip)
}

// RegisterIPUpdate 通知网关更新
func RegisterIPUpdate(ctx *protocol.HandlerCtx) {
	arg := &entity.RegisterIpUpdateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	logger.Info("通知网关更新 = ", arg.Project, arg.TargetPort)
	gateway.GetIps(arg.Project, arg.TargetPort)
}

// GatewayDel 停止并删除网关
func GatewayDel(ctx *protocol.HandlerCtx) {
	project := ""
	err := protocol.DataDecoder(ctx.Stream.Data, &project)
	if err != nil {
		logger.Error("docker pull err  = ", err)
	}
	logger.Info("停止并删除网关 : ", project)
	gateway.CloseGatewayServer(project)
	buf, err := protocol.DataEncoder("网关已经启动")
	if err != nil {
		fmt.Println(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		fmt.Println(err)
	}
}

// SlaveHosts 读取 hosts文件
func SlaveHosts(ctx *protocol.HandlerCtx) {
	hostsData := cmd2.LinuxSendCommand("cat /etc/hosts")
	log.Println("读取 hosts文件 : ", hostsData)
	rse := &entity.SlaveHostsRse{
		Data: hostsData,
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// SlaveHostsUpdate 修改 hosts文件
func SlaveHostsUpdate(ctx *protocol.HandlerCtx) {
	arg := &entity.SlaveHostsArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}

	// 覆盖文件内容
	// 打开一个存在的文件，将原来的内容覆盖掉
	path := "/etc/hosts"
	// O_WRONLY: 只写, O_TRUNC: 清空文件
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("文件打开错误", err)
		return
	}

	defer file.Close() // 关闭文件
	// 带缓冲区的*Writer
	writer := bufio.NewWriter(file)
	writer.WriteString(arg.Data)
	// 将缓冲区中的内容写入到文件里
	writer.Flush()

	SlaveHosts(ctx)
}

// GetSlavePathInfo 获取指定路径的目录结构与文件
func GetSlavePathInfo(ctx *protocol.HandlerCtx) {
	arg := &entity.GetSlavePathInfoArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}

	//logger.Info("获取目录结构 =  ", arg.Path)

	rse := &entity.GetSlavePathInfoRse{
		Error:         "",
		Rse:           "成功",
		TaskId:        arg.TaskId,
		FileStructure: make([]*entity.FileStructure, 0),
	}

	// 获取目录结构
	files, err := ioutil.ReadDir(arg.Path)
	if err != nil {
		log.Println(err)
		rse.Error = err.Error()
		rse.Rse = "获取失败"
		goto R
	}

	for _, v := range files {
		filename := v.Name()
		isDir := v.IsDir()
		fileNameWithSuffix := path.Base(filename)
		//获取文件的后缀(文件类型)
		isEdit := false
		fileType := path.Ext(fileNameWithSuffix)
		if fileType == ".conf" || fileType == ".yaml" || fileType == ".ini" || fileType == ".txt" ||
			fileType == ".json" {
			isEdit = true
		}
		isZip := false
		if fileType == ".zip" || fileType == ".gz" || fileType == ".7z" || fileType == ".rar" ||
			fileType == ".ar" || fileType == ".apz" || fileType == ".bz2" {
			isZip = true
		}
		rse.FileStructure = append(rse.FileStructure, &entity.FileStructure{
			FileName: filename,
			IsDir:    isDir,
			IsEdit:   isEdit,
			IsZip:    isZip,
		})
		//log.Println(filename, isDir, isEdit)
		//log.Println(v)
		//log.Println("_______________________")
	}

R:
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}

}

// SlaveCatFile 查看指定文件内容
func SlaveCatFile(ctx *protocol.HandlerCtx) {
	arg := &entity.CatSlaveFileArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}

	//logger.Info("指定文件路径 =  ", arg.FilePath)

	note := cmd2.LinuxSendCommand("cat " + arg.FilePath)
	//logger.Println("读取文件 : ", note)
	rse := &entity.CatSlaveFileRse{
		Data: note,
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// SlaveMkdir 创建目录
func SlaveMkdir(ctx *protocol.HandlerCtx) {
	arg := &entity.SlaveMkdirArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	rse := &entity.SlaveMkdirRse{
		TaskId: arg.TaskId,
	}
	err = os.Mkdir(arg.Path, 0666)
	if err != nil {
		rse.Rse = err.Error()
	} else {
		rse.Rse = "成功"
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// SlaveDecompress 解压文件
func SlaveDecompress(ctx *protocol.HandlerCtx) {
	arg := &entity.SlaveDecompressArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	rse := &entity.SlaveDecompressRse{
		TaskId: arg.TaskId,
	}
	dst := arg.Path
	fileType := path.Ext(arg.Path)
	outList := strings.Split(dst, ".")
	outPath := strings.Join(outList[0:len(outList)-1], ".") + "/"

	if !(fileType == ".zip" || fileType == ".tar" || fileType == ".gz") {
		rse.Rse = "不支持的文件类型"
		goto R
	}
	// 解压文件
	logger.Info("解压文件 : ", dst, " -> ", outPath)
	if fileType == ".zip" {
		err = utils.DeCompressZIP(dst, outPath)
	} else if fileType == ".tar" || fileType == ".gz" {
		err = utils.DeCompressTAR(dst, outPath)
	}
R:
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// NginxInfo 获取nginx 信息
func NginxInfo(ctx *protocol.HandlerCtx) {
	arg := &entity.NginxInfoArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	rse := &entity.NginxInfoRse{
		TaskId: arg.TaskId,
	}
	info := &entity.NginxInfo{}

	_, info.Version = linux.HasNginx()
	info.ConfPath = linux.NginxConfPath()
	info.LogPath = "/var/log/nginx"
	info.PID = linux.NginxPid()
	info.Status = linux.NginxServiceStatus()
	rse.Rse = info

	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// NginxStart 启动nginx
func NginxStart(ctx *protocol.HandlerCtx) {
	arg := &entity.NginxStartArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	rse := &entity.NginxStartRse{
		TaskId: arg.TaskId,
	}
	txt := linux.NginxStart()
	if len(txt) == 0 {
		rse.Rse = "成功"
	} else {
		rse.Rse = txt
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// NginxReload 重启nginx
func NginxReload(ctx *protocol.HandlerCtx) {
	arg := &entity.NginxReloadArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	rse := &entity.NginxReloadRse{
		TaskId: arg.TaskId,
	}
	txt := linux.NginxReload()
	if len(txt) == 0 {
		rse.Rse = "成功"
	} else {
		rse.Rse = txt
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// NginxQuit 停止nginx
func NginxQuit(ctx *protocol.HandlerCtx) {
	arg := &entity.NginxQuitArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	rse := &entity.NginxQuitRse{
		TaskId: arg.TaskId,
	}
	txt := linux.NginxQuit()
	if len(txt) == 0 {
		rse.Rse = "成功"
	} else {
		rse.Rse = txt
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// NginxStop 强制停止nginx
func NginxStop(ctx *protocol.HandlerCtx) {
	arg := &entity.NginxStopArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	rse := &entity.NginxStopRse{
		TaskId: arg.TaskId,
	}
	txt := linux.NginxStop()
	if len(txt) == 0 {
		rse.Rse = "成功"
	} else {
		rse.Rse = txt
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// NginxCheckConf 检查配置
func NginxCheckConf(ctx *protocol.HandlerCtx) {
	arg := &entity.NginxCheckConfArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	rse := &entity.NginxCheckConfRse{
		TaskId: arg.TaskId,
		Rse:    linux.NginxT(),
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// NginxConfUpdate 修改nginx 配置文件
func NginxConfUpdate(ctx *protocol.HandlerCtx) {
	arg := &entity.NginxConfUpdateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	logger.Info("修改nginx配置文件 path = ", arg.Path)
	rse := &entity.NginxCheckConfRse{
		TaskId: arg.TaskId,
	}
	var writer *bufio.Writer
	// 覆盖文件内容
	// 打开一个存在的文件，将原来的内容覆盖掉
	// O_WRONLY: 只写, O_TRUNC: 清空文件
	file, err := os.OpenFile(arg.Path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		logger.Error("文件打开错误", err)
		rse.Rse = "文件打开错误"
		goto R
	}
	defer file.Close() // 关闭文件
	// 带缓冲区的*Writer
	writer = bufio.NewWriter(file)
	_, err = writer.WriteString(arg.Data)
	if err != nil {
		rse.Rse = err.Error()
		goto R
	}
	// 将缓冲区中的内容写入到文件里
	err = writer.Flush()
	if err != nil {
		rse.Rse = err.Error()
		goto R
	}
	rse.Rse = "修改成功"
R:
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}
