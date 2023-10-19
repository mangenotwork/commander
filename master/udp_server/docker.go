package udp_server

import (
	"fmt"
	"strings"
	"time"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/enum"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"
	"gitee.com/mangenotework/commander/master/handler"

	"github.com/docker/docker/api/types"
)

func DockerInfo(ctx *HandlerCtx) {
	info := types.Info{}
	//err := protocol.GobDecoder(info, ctx.Stream.Data)
	err := protocol.DataDecoder(ctx.Stream.Data, &info)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- info
	}
	return
}

func DockerPS(ctx *HandlerCtx) {
	data := make([]types.Container, 0)
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	if err != nil {
		logger.Error("protocol.GobDecoder err = ", err)
	}
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data
	}
	return
}

func DockerImages(ctx *HandlerCtx) {
	data := make([]types.ImageSummary, 0)
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

func DockerPull(ctx *HandlerCtx) {
	data := &entity.DockerPullResult{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	if err != nil {
		data.Err = err
	}
	logger.Info("DockerPull data = ", data)
	task, err := new(dao.DaoTask).Get(data.TaskId)
	if err != nil || task == nil {
		// 下发通知
		handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker pull" +
			" 失败， task 为空" +
			"; " + utils.StringValue(data)))
		return
	}
	if data.Err != nil {
		task.State = enum.TaskStateBreakOff.Value()
		task.StateStr = enum.TaskStateBreakOff.Str()
		task.Result = data.Err.Error()
	} else {
		task.State = enum.TaskStateComplete.Value()
		task.StateStr = enum.TaskStateComplete.Str()
		task.Result = data
	}
	_ = new(dao.DaoTask).Set(data.TaskId, task)
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker pull " +
		task.StateStr + "; " + utils.StringValue(data)))
	return
}

func DockerRun(ctx *HandlerCtx) {
	data := &entity.DockerRunResult{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	if err != nil {
		data.Err = err
	}
	logger.Info("DockerRun data = ", data)
	task, err := new(dao.DaoTask).Get(data.TaskId)
	if err != nil || task == nil {
		// 下发通知
		handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker pull" +
			" 失败， task 为空" +
			"; " + utils.StringValue(data)))
		return
	}
	if data.Err != nil {
		task.State = enum.TaskStateBreakOff.Value()
		task.StateStr = enum.TaskStateBreakOff.Str()
		task.Result = data.Err.Error()
	} else {
		task.State = enum.TaskStateComplete.Value()
		task.StateStr = enum.TaskStateComplete.Str()
		task.Result = data
	}
	_ = new(dao.DaoTask).Set(data.TaskId, task)
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker run " +
		task.StateStr + "; " + utils.StringValue(data)))
	// 如果是项目的部署容器 则需要保存数据
	if data.IsProject {
		err = new(dao.DaoProjectDocker).SetProjectDockerContainer(data.Project, data.ID, &entity.DockerContainerDeploy{
			Slave:       ctx.IP,
			Project:     data.Project,
			ContainerId: data.ID,
			TaskID:      task.ID,
			Port:        data.Port,
		})
		if err != nil {
			logger.Error("保存项目的部署容器数据失败 err = ", err.Error())
		}
		// 通知网关更新
		projectObj, pErr := new(dao.DaoProjectDocker).Get(data.Project)
		if pErr != nil {
			logger.Error("获取项目对象失败, err = ", pErr.Error())
			return
		}
		if projectObj.IsGateway == "1" {
			for k, v := range data.Port {
				arg := &entity.RegisterIpToGatewayArg{
					Key: fmt.Sprintf("%s%s", data.Project, v),
					Ip:  ctx.IP + ":" + k,
				}
				logger.Error(arg.Key, arg.Ip)
				buf, err := protocol.DataEncoder(arg)
				if err != nil {
					logger.Error("在网关上注册地址失败")
					return
				}
				rse, err := UDPSend(projectObj.GatewaySlave, protocol.CMD_RegisterIpToGateway, buf)
				if err != nil {
					logger.Error(err)
				}
				logger.Info(rse)
			}
		}
		// 更新项目副本数量到持久化
		_, keys, err := new(dao.DaoProjectDocker).GetProjectDockerContainer(data.Project)
		if err != nil {
			logger.Error(err)
			return
		}
		project, err := new(dao.DaoProjectDocker).Get(data.Project)
		if err != nil {
			logger.Error(err)
			return
		}
		project.Duplicate = utils.StringValue(len(keys))
		project.UpdateTime = utils.NowTimeStr()
		err = new(dao.DaoProjectDocker).Set(data.Project, project)
		if err != nil {
			logger.Error(err)
			return
		}
	}
	return
}

func UDPSend(slave string, cmd protocol.CommandCode, data []byte) (string, error) {
	udpC, ok := protocol.AllUdpClient.RetryGet(slave)
	if !ok {
		logger.Infof("192.168.0.9  离线 ok is false")
	}
	logger.Infof("udpC = ", udpC)
	if udpC == nil {
		logger.Info("192.168.0.9  离线  udpC = nil ")
		return "", fmt.Errorf("udp is null")
	}
	requst := utils.IDMd5() // 6854823418404110336
	logger.Infof("requst = ", requst)
	packate, err := protocol.Packet(cmd, requst, data)
	if err != nil {
		logger.Error(err)
	}
	logger.Infof("发送数据: ", packate)
	logger.Info(udpC.Conn, udpC.IP)
	protocol.UDPSend(udpC.Conn, packate)
	protocol.Set(requst)
	return requst, nil
}

func DockerStop(ctx *HandlerCtx) {
	data := &entity.DockerStopRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	if err != nil {
		logger.Error(err)
	}
	// 任务记录
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId, data.Rse)
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker stop " +
		data.ContainerId + "; " + data.Rse))
	return
}

// DockerRm docker rm
func DockerRm(ctx *HandlerCtx) {
	data := &entity.DockerContainerRmRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	if err != nil {
		logger.Error(err)
	}
	// 任务记录
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId, data.Rse)
	// 如果是项目的则需要修改数据
	if data.IsProject == "1" {
		err = new(dao.DaoProjectDocker).DelProjectDockerContainer(data.Project, data.ContainerId)
		if err != nil {
			logger.Error(err)
		}
		// 获取项目信息
		projectObj, pErr := new(dao.DaoProjectDocker).Get(data.Project)
		if pErr != nil {
			logger.Error(pErr)
		}
		// 通知网关更新
		if projectObj.IsGateway == "1" {
			for _, v := range strings.Split(projectObj.Port, ";") {
				arg := &entity.RegisterIpUpdateArg{
					Project:    data.Project,
					TargetPort: strings.Split(v, ":")[1],
				}
				logger.Error("通知更新网关ips = ", arg.Project, arg.TargetPort)
				buf, err := protocol.DataEncoder(arg)
				if err != nil {
					logger.Error("在网关上注册地址失败")
					return
				}
				rse, err := UDPSend(data.Slave, protocol.CMD_RegisterIPUpdate, buf)
				if err != nil {
					logger.Error(err)
				}
				logger.Info(rse)
			}
		}
		// 更新项目持久化
		_, keys, err := new(dao.DaoProjectDocker).GetProjectDockerContainer(data.Project)
		if err != nil {
			logger.Error(err)
			return
		}
		project, err := new(dao.DaoProjectDocker).Get(data.Project)
		if err != nil {
			logger.Error(err)
			return
		}
		project.Duplicate = utils.StringValue(len(keys))
		project.UpdateTime = utils.NowTimeStr()
		err = new(dao.DaoProjectDocker).Set(data.Project, project)
		if err != nil {
			logger.Error(err)
			return
		}
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker rm " +
		data.ContainerId + "; " + data.Rse))
	//tx, err := protocol.Get(ctx.Stream.CtxId)
	//if err == nil {
	//	tx.Data <- string(ctx.Stream.Data)
	//}
	return
}

// DockerRmi docker rmi
func DockerRmi(ctx *HandlerCtx) {
	data := &entity.DockerRmiRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	if err != nil {
		logger.Error(err)
	}
	// 任务记录
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId, data.Rse)
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker rmi " +
		data.ImageId + "; " + data.Rse))
	return
}

// ContainerLog container log
func ContainerLog(ctx *HandlerCtx) {
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- string(ctx.Stream.Data)
	}
	return
}

func ExecutablePIDLog(ctx *HandlerCtx) {
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- string(ctx.Stream.Data)
	}
	return
}

func ExecutableKill(ctx *HandlerCtx) {
	data := &entity.ExecutableKillRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	task, taskErr := new(dao.DaoSlaveExecutableTask).Get(data.TaskId)
	if taskErr != nil {
		logger.Error("task get err : ", taskErr)
		return
	}
	task.State = enum.ExecutableStateDiscontinued
	task.PID = 0
	taskErr = new(dao.DaoSlaveExecutableTask).Set(data.TaskId, task)
	if taskErr != nil {
		logger.Error("task set err : ", taskErr)
		return
	}
	// 记录任务
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId, data.Rse)
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
}

func ContainerTop(ctx *HandlerCtx) {
	data := &entity.ContainerTopResult{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

func SlaveProcessInfo(ctx *HandlerCtx) {
	data := &entity.ProcessInfo{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

func ContainerRename(ctx *HandlerCtx) {
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- string(ctx.Stream.Data)
	}
	return
}

// ContainerRestart docker restart
func ContainerRestart(ctx *HandlerCtx) {
	data := &entity.DockerRestartRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 任务记录
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId, data.Rse)
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker restart " +
		data.ContainerId + "; " + data.Rse))
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data.Rse
	}
	return
}

func ContainerPause(ctx *HandlerCtx) {
	data := &entity.DockerPauseRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 任务记录
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId, data.Rse)
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: docker pause " +
		data.ContainerId + "; " + data.Rse))
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data.Rse
	}
	return
}

func DockerStateS(ctx *HandlerCtx) {
	data := make([]*entity.ContainerPerformance, 0)
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

func DockerImageRun(ctx *HandlerCtx) {
	data := &entity.DockerImageRunRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 任务记录
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId, data.Rse)
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: 部署结果 " + data.Rse))
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data.Rse
	}
	return
}

func SlaveHosts(ctx *HandlerCtx) {
	data := &entity.SlaveHostsRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data.Data
	}
	return
}

func SlaveHostsUpdate(ctx *HandlerCtx) {
	data := &entity.SlaveHostsRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data.Data
	}
	return
}

// WarningNotice 收到来自slave的警告通知
func WarningNotice(ctx *HandlerCtx) {
	notice := string(ctx.Stream.Data)
	//持久化报警
	id := utils.IDStr()
	_ = new(dao.DaoMonitorRule).SetAlarm(id, &entity.Alarm{
		ID:    id,
		Slave: ctx.IP,
		Date:  utils.NowTimeStr(),
		Note:  fmt.Sprintf("可执行文件停止运行: ", notice),
		Lv:    "警报",
	})
	// 下发通知
	handler.BroadcastNotice([]byte(utils.NowTimeStr() + " 可执行文件停止运行: " + notice))
}

// EnvDeployedCheck 检查是否安装指定软件
func EnvDeployedCheck(ctx *HandlerCtx) {
	data := &entity.EnvDeployedCheckRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data
	}
	return
}

// InstallDocker 安装docker 返回的信息
func InstallDocker(ctx *HandlerCtx) {
	data := &entity.InstallDockerRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data
	}
	return
}

// RemoveDocker 卸载docker 返回的信息
func RemoveDocker(ctx *HandlerCtx) {
	data := &entity.RemoveDockerRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data
	}
	return
}

// InstallNginx 安装nginx 返回的信息
func InstallNginx(ctx *HandlerCtx) {
	data := &entity.InstallNginxRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data
	}
	return
}

// RemoveNginx 卸载nginx 返回的信息
func RemoveNginx(ctx *HandlerCtx) {
	data := &entity.RemoveNginxRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	// 输出到http
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- data
	}
	return
}
