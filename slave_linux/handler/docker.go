package handler

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"gitee.com/mangenotework/commander/common/docker"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/docker/docker/api/types/container"
)

// DockerInfo docker 信息
func DockerInfo(ctx *protocol.HandlerCtx) {
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		return
	}
	info, _ := d.DockerInfo()
	buf, err := protocol.DataEncoder(info)
	if err != nil {
		logger.Error(err)
		return
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// DockerPs docker 查看所有 容器
func DockerPs(ctx *protocol.HandlerCtx) {
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		return
	}
	data, err := d.DockerPs()
	if err != nil {
		logger.Error("DockerPs err = ", err)
		return
	}
	buf, err := protocol.DataEncoder(data)
	logger.Info("protocol.Struct2Byte err = ", err)
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// DockerImages docker 查看所有 镜像
func DockerImages(ctx *protocol.HandlerCtx) {
	logger.Info("DockerImages")
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		return
	}
	data, err := d.DockerImages()
	if err != nil {
		logger.Error("DockerPs err = ", err)
		return
	}
	buf, err := protocol.DataEncoder(data)
	if err != nil {
		logger.Error(err)
		return
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// DockerPull docker 拉取镜像, 返回任务ID
// TODO: 添加一个锁， 当前一个节点只能执行一个pull, 任务ID查看进度和结果
func DockerPull(ctx *protocol.HandlerCtx) {
	var (
		d      *docker.DockerClient
		data   []byte
		pull   = entity.DockerPullArg{}
		result = &entity.DockerPullResult{
			Err: nil,
		}
	)
	err := protocol.DataDecoder(ctx.Stream.Data, &pull)
	if err != nil {
		logger.Error("docker pull err  = ", err)
		result.Err = err
		goto Send
	}

	logger.Info("pull = ", pull)
	result.TaskId = pull.TaskId
	d, err = docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		result.Err = err
		goto Send
	}

	data, err = d.DockerPull(pull)
	if err != nil {
		logger.Error("DockerPull err = ", err)
		result.Err = err
		goto Send
	}

Send:
	logger.Info("DockerPull data = ", data, result.TaskId)
	result.Data = string(data)
	buf, err := protocol.DataEncoder(result)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// DockerRun docker 启动一个容器
func DockerRun(ctx *protocol.HandlerCtx) {
	var (
		d    *docker.DockerClient
		arg  = entity.DockerRunArg{}
		data container.ContainerCreateCreatedBody
		obj  = entity.DockerRunResult{}
		port = map[string]string{}
		info = []byte{}
	)

	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("docker run err  = ", err)
		obj.Err = err
		goto Send
	}
	logger.Info("arg = ", arg)

	if arg.RandomPort == 1 {
		// 随机一个5位数， 并且查看这个5位数作为端口是否被占用
		for _, v := range arg.PortRelation {
			port[randPort()] = v
		}
	} else {
		for k, v := range arg.PortRelation {
			port[k] = v
		}
	}

	obj.TaskId = arg.TaskId
	d, err = docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		obj.Err = err
		goto Send
	}
	logger.Info("arg.ImagePass = ", arg.ImagePass)
	info, err = d.DockerPull(entity.DockerPullArg{
		Name:   arg.ImageUser,
		Pass:   arg.ImagePass,
		Image:  arg.Image,
		TaskId: arg.TaskId,
	})
	logger.Info("拉取镜像 , info = ", string(info), err)

	arg.PortRelation = port
	data, err = d.DockerRun(arg)
	if err != nil {
		logger.Error("docker run err  = ", err)
		obj.Err = err
		goto Send
	}
Send:
	logger.Info("data = ", data)
	obj.ID = data.ID
	obj.Warnings = data.Warnings
	obj.IsProject = arg.IsProject
	obj.Project = arg.Project
	obj.Port = port
	buf, err := protocol.DataEncoder(obj)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

func randPort() string {
	rand.Seed(time.Now().UnixNano())
	for {
		i := rand.Intn(56000)
		if i < 10000 {
			i += 10000
		}
		logger.Info("端口  i= ", i)
		data := LinuxSendCommand(fmt.Sprintf("netstat -anp | grep %d", i))
		logger.Info(data, len(data))
		if len(data) == 0 {
			return utils.StringValue(i)
		}
	}
}

// DockerStop docker 关闭一个容器
func DockerStop(ctx *protocol.HandlerCtx) {
	arg := &entity.DockerStopArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		log.Println("docker pull err  = ", err)
	}

	logger.Info("containerId = ", arg.ContainerId)
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Info("本地docker客户端 err = ", err)
		_ = ctx.Send([]byte("关闭容器失败: " + err.Error()))
		return
	}
	err = d.DockerStop(arg.ContainerId)

	rse := &entity.DockerStopRse{
		TaskId:      arg.TaskId,
		ContainerId: arg.ContainerId,
	}
	if err != nil {
		rse.Rse = "关闭容器失败: " + err.Error()
	} else {
		rse.Rse = "关闭容器成功"
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

// DockerRm docker 删除一个容器
func DockerRm(ctx *protocol.HandlerCtx) {
	arg := &entity.DockerContainerRmArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		log.Println("docker pull err  = ", err)
	}

	logger.Info("containerId = ", arg.ContainerId)
	logger.Info("删除一个容器 ......")
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		_ = ctx.Send([]byte("删除容器失败: " + err.Error()))
		return
	}
	err = d.DockerRm(arg.ContainerId)
	if err != nil {
		logger.Error("删除容器失败 err = ", err)
		_ = ctx.Send([]byte("删除容器失败: " + err.Error()))
		return
	}
	logger.Info("删除容器成功")

	rse := &entity.DockerContainerRmRse{
		IsProject:   arg.IsProject,
		Slave:       arg.Slave,
		Project:     arg.Project,
		ContainerId: arg.ContainerId,
		TaskId:      arg.TaskID,
		Rse:         "删除容器成功",
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

// DockerRmi docker 删除一个镜像
func DockerRmi(ctx *protocol.HandlerCtx) {
	arg := &entity.DockerRmiArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		log.Println("docker pull err  = ", err)
	}

	logger.Info("imageId = ", arg.ImageId)
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		_ = ctx.Send([]byte("删除镜像失败: " + err.Error()))
		return
	}
	_, err = d.DockerRmi(arg.ImageId)

	rse := &entity.DockerRmiRse{
		TaskId:  arg.TaskId,
		ImageId: arg.ImageId,
	}

	if err != nil {
		logger.Error("删除镜像失败 err = ", err)
		rse.Rse = "删除镜像失败 err = " + err.Error()
	} else {
		rse.Rse = "删除容器成功"
	}

	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
	return
}

type ContainerLogIO struct {
	IO   io.ReadCloser
	Stop chan int
}

var ContainerLogIOMap = make(map[string]*ContainerLogIO)

func CloseAllContainerLogIO() {
	logger.Info("CloseAllContainerLogIO .......")
	for _, v := range ContainerLogIOMap {
		logger.Info(v)
		logger.Info("Close ContainerLogIO ......")
		v.Stop <- 1
	}
}

// ContainerLog docker 查看容器日志
func ContainerLog(ctx *protocol.HandlerCtx) {
	containerId := string(ctx.Stream.Data)
	logger.Info("containerId = ", containerId)
	if containerId != "close" {
		d, err := docker.NewDockerClient()
		if err != nil {
			logger.Error("本地docker客户端 err = ", err)
			_ = ctx.Send([]byte("删除容器失败: " + err.Error()))
			return
		}
		data, err := d.ContainerLog(containerId)
		if err != nil {
			logger.Error(err)
			return
		}
		logIO := &ContainerLogIO{
			IO:   data,
			Stop: make(chan int),
		}
		ContainerLogIOMap[ctx.Stream.CtxId] = logIO

		go func() {
			buf := make([]byte, 1024)
			for {
				n, errRead := logIO.IO.Read(buf)
				if errRead != nil && errRead.Error() == "http: read on closed response body" {
					logger.Error(errRead)
					return
				}
				if n > 0 && errRead == nil {
					logger.Info(string(buf))
					_ = ctx.Send(buf)
				}
				// 解决 避免cpu 拉太高
				// 解决 unix /var/run/docker.sock: socket: too many open files
				time.Sleep(3000 * time.Millisecond) // 3秒读一次
			}
		}()

		for {
			select {
			case <-logIO.Stop:
				_ = logIO.IO.Close()
				logger.Info("logIO.Stop ...... ")
				return
			}
		}

	} else {
		logger.Info("ContainerLog  Close ...... ")
		logIO, ok := ContainerLogIOMap[ctx.Stream.CtxId]
		if ok {
			logIO.Stop <- 1
		}
	}
}

// ContainerTop 查看容器进程
func ContainerTop(ctx *protocol.HandlerCtx) {
	obj := &entity.ContainerTopResult{}
	containerId := string(ctx.Stream.Data)
	logger.Info("containerId = ", containerId)
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("err = ", err)
		obj.Err = err
		goto Send
	}
	obj.Data, err = d.ContainerTop(containerId)
	if err != nil {
		obj.Err = err
		goto Send
	}
Send:
	buf, err := protocol.DataEncoder(obj)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// ContainerRename 修改容器名称
func ContainerRename(ctx *protocol.HandlerCtx) {
	arg := &entity.ContainerReNameArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("err = ", err)
		_ = ctx.Send([]byte("修改失败 : " + err.Error()))
		return
	}
	err = d.ContainerRename(arg.ContainerId, arg.NewName)
	if err != nil {
		_ = ctx.Send([]byte("修改失败 : " + err.Error()))
		return
	}
	_ = ctx.Send([]byte("修改成功"))
	return
}

// ContainerRestart 容器重启
func ContainerRestart(ctx *protocol.HandlerCtx) {
	arg := &entity.DockerRestartArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	logger.Info("containerId = ", arg.ContainerId)
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		_ = ctx.Send([]byte("重启容器失败: " + err.Error()))
		return
	}
	rse := &entity.DockerRestartRse{
		TaskId:      arg.TaskId,
		ContainerId: arg.ContainerId,
	}
	err = d.ContainerRestart(arg.ContainerId)
	if err != nil {
		logger.Error("容器重启失败 err = ", err)
		rse.Rse = "容器重启 err = " + err.Error()
	} else {
		rse.Rse = "容器重启成功"
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
	return
}

// ContainerPause 容器暂停
func ContainerPause(ctx *protocol.HandlerCtx) {
	arg := &entity.DockerPauseArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	logger.Info("containerId = ", arg.ContainerId)
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		_ = ctx.Send([]byte("暂停容器失败: " + err.Error()))
		return
	}
	err = d.ContainerPause(arg.ContainerId)
	rse := &entity.DockerPauseRse{
		TaskId:      arg.TaskId,
		ContainerId: arg.ContainerId,
	}
	err = d.ContainerRestart(arg.ContainerId)
	if err != nil {
		logger.Error("暂停容器失败 err = ", err)
		rse.Rse = "暂停容器失败 err = " + err.Error()
	} else {
		rse.Rse = "暂停容器成功"
	}
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
	return
}

// DockerStateS 容器资源使用情况
func DockerStateS(ctx *protocol.HandlerCtx) {
	containerId := string(ctx.Stream.Data)
	//logger.Info("containerId = ", containerId)
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		_ = ctx.Send([]byte("暂停容器失败: " + err.Error()))
		return
	}

	containerIdList := strings.Split(containerId, ",")
	rse := make([]*entity.ContainerPerformance, 0)
	var wg sync.WaitGroup
	for _, id := range containerIdList {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			//logger.Info("idididididid  = ", id)
			data := d.DockerState(id)
			rse = append(rse, data)
		}(id)
	}
	wg.Wait()

	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// DockerImageRun 运行指定镜像
// 业务逻辑
// 1. 检查镜像是否存在
// 2. 如果存在则执行
func DockerImageRun(ctx *protocol.HandlerCtx) {
	arg := &entity.DockerImageRunArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	d, err := docker.NewDockerClient()
	if err != nil {
		logger.Error("本地docker客户端 err = ", err)
		_ = ctx.Send([]byte("暂停容器失败: " + err.Error()))
		return
	}
	imgs, err := d.DockerImages()
	if err != nil {
		logger.Error("获取镜像列表失败: ", err.Error())
		return
	}
	flage := false
	for _, v := range imgs {
		if arg.ImageId == v.ID {
			flage = true
			break
		}
	}
	port := map[string]string{}
	if arg.RandomPort == 1 {
		// 随机一个5位数， 并且查看这个5位数作为端口是否被占用
		for _, v := range arg.PortRelation {
			port[randPort()] = v
		}
	} else {
		for k, v := range arg.PortRelation {
			port[k] = v
		}
	}
	rse := &entity.DockerImageRunRse{
		TaskId: arg.TaskId,
	}
	if flage {
		_, err = d.DockerRun(entity.DockerRunArg{
			Image:        arg.ImageId,
			RandomPort:   arg.RandomPort,
			PortRelation: port,
			Name:         arg.Name,
			Env:          arg.Env,
			TaskId:       arg.TaskId,
			IsProject:    arg.IsProject,
			Project:      arg.Project,
			IsAlways:     arg.IsAlways,
		})
		if err != nil {
			rse.Rse = "部署镜像失败: " + err.Error()
		} else {
			rse.Rse = "部署镜像成功"
		}
	} else {
		rse.Rse = "部署镜像失败: 没有该镜像"
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
