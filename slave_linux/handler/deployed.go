package handler

import (
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/slve_linux/linux"
)

// EnvDeployedCheck 检查软件是否安装
func EnvDeployedCheck(ctx *protocol.HandlerCtx) {
	arg := &entity.EnvDeployedCheckArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}

	rse := &entity.EnvDeployedCheckRse{}
	softwareList := make([]*entity.SoftwareDeployedCheck, 0)

	for _, v := range arg.Software {
		if v == "docker" {
			logger.Info("是否安装docker...")
			is, info := linux.HaveDocker()
			if is {
				info = linux.CmdDockerVersion()
			}
			logger.Info("docker info = ", info)
			softwareList = append(softwareList, &entity.SoftwareDeployedCheck{
				Software: v,
				IsHave:   is,
				Info:     info,
			})
		}
		if v == "nginx" {
			logger.Info("是否安装nginx...")
			is, info := linux.HasNginx()
			logger.Info("nginx info = ", info)
			softwareList = append(softwareList, &entity.SoftwareDeployedCheck{
				Software: v,
				IsHave:   is,
				Info:     info,
			})
		}
	}
	rse.SoftwareCheck = softwareList
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// InstallDocker 安装 docker
func InstallDocker(ctx *protocol.HandlerCtx) {
	info := linux.DeployedDockerCE()
	rse := &entity.InstallDockerRse{
		Rse: info,
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

// RemoveDocker 卸载 docker
func RemoveDocker(ctx *protocol.HandlerCtx) {
	info := linux.RemoveDocker()
	rse := &entity.RemoveDockerRse{
		Rse: info,
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

// InstallNginx 安装 nginx
func InstallNginx(ctx *protocol.HandlerCtx) {
	info := linux.DeployedNginx()
	rse := &entity.RemoveDockerRse{
		Rse: info,
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

// RemoveNginx 卸载 nginx
func RemoveNginx(ctx *protocol.HandlerCtx) {
	info := linux.RemoveNginx()
	rse := &entity.RemoveDockerRse{
		Rse: info,
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
