package handler

import (
	"fmt"
	"strings"
	"time"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

// ProjectDockerCreate 创建一个项目
func ProjectDockerCreate(c *gin.Context) {
	dockerName := c.Request.PostFormValue("docker_name")            // 项目的名称 唯一的
	dockerNote := c.Request.PostFormValue("docker_note")            // 项目的备注
	dockerIsGateway := c.Request.PostFormValue("docker_is_gateway") // 是否启动一个网关
	dockerPort := c.Request.PostFormValue("docker_port")            // 项目的端口映射
	dockerImage := c.Request.PostFormValue("docker_image")          // 项目镜像
	dockerUser := c.Request.PostFormValue("docker_user")            // 项目拉取镜像的账号
	dockerPassword := c.Request.PostFormValue("docker_password")    // 项目拉取镜像的密码
	dockerEnv := c.Request.PostFormValue("docker_env")              // 启动容器的环境变量
	dockerVolume := c.Request.PostFormValue("docker_volume")        // 容器的 volume
	dockerDuplicate := c.Request.PostFormValue("docker_duplicate")  // 启动容器的个数
	gatewayPort := c.Request.PostFormValue("gateway_port")          // 网关端口映射
	gatewaySlave := c.Request.PostFormValue("gateway_slave")        // 网关部署在哪个主机上
	if len(dockerName) == 0 || len(dockerImage) == 0 || len(dockerDuplicate) == 0 {
		APIOutPut(c, 1, 0, "", "参数错误")
		return
	}

	p, _ := new(dao.DaoProjectDocker).Get(dockerName)
	if p != nil {
		APIOutPut(c, 1, 0, "", "项目已经存在，请重命名项目，可以已版本号区分。")
		return
	}

	projectDocker := &entity.ProjectDocker{
		Name:         dockerName,
		Note:         dockerNote,
		IsGateway:    dockerIsGateway,
		Port:         dockerPort,
		Image:        dockerImage,
		User:         dockerUser,
		Password:     dockerPassword,
		Env:          dockerEnv,
		Volume:       dockerVolume,
		Duplicate:    dockerDuplicate,
		CreateTime:   utils.NowTimeStr(),
		GatewaySlave: gatewaySlave,
		GatewayPort:  gatewayPort,
	}
	err := new(dao.DaoProjectDocker).Set(dockerName, projectDocker)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}

	// 部署并运行
	// 获取当前在线的主机， 部署满指定个数
	online := protocol.AllUdpClient.GetAllKey()
	n := 0
	l := len(online) - 1
	i := 0
	for i < utils.Num2Int(dockerDuplicate) {
		if n >= l {
			n = 0
		}
		deployErr := deployDocker(projectDocker, online[n])
		if deployErr == nil {
			i++
		}
		n++
	}
	// 如果设置了网关服务则启动网关服务
	if dockerIsGateway == "1" {
		deployGateway(dockerName, gatewayPort, gatewaySlave)
	}
	APIOutPut(c, 1, 0, "ok", "ok")
}

// deployGateway 部署网关服务
func deployGateway(project, gatewayPort, gatewaySlave string) {
	go func() {
		arg := &entity.GatewayArg{
			Ports:       strings.Split(gatewayPort, ";"),
			ProjectName: project,
			LVS:         "L4 - TCP",
			LVSModel:    "random",
		}
		buf, err := protocol.DataEncoder(arg)
		if err != nil {
			logger.Error("启动网关失败")
			return
		}
		rse, err := UDPSend(gatewaySlave, protocol.CMD_GatewayRun, buf)
		if err != nil {
			logger.Error(err)
		}
		logger.Info(rse)
		// 记录网关数据，持久化
		gatewayBase := &entity.GatewayBase{
			Slave:       gatewaySlave,
			Ports:       arg.Ports,
			ProjectName: arg.ProjectName,
			LVS:         arg.LVS,
			LVSModel:    arg.LVSModel,
			IsClose:     "0",
			Create:      utils.NowTimeStr(),
		}
		err = new(dao.DaoGateway).Set(project, gatewayBase)
		if err != nil {
			logger.Error("记录网关数据，持久化失败, err = ", err)
		}
		// 启动一个协成监控网关服务
		for _, v := range arg.Ports {
			portList := strings.Split(v, ":")
			if len(portList) < 1 {
				continue
			}
			port := portList[0]
			go func() {
				for {
					if !utils.Tcper(gatewaySlave + ":" + port) {
						gatewayObj, pErr := new(dao.DaoGateway).Get(project)
						if pErr != nil {
							continue
						}
						flag := false
						// 更新port的情况
						for _, v := range gatewayObj.Ports {
							if strings.Split(v, ":")[0] == port { // 如果更新了就找不到，也不会被监控到
								if gatewayObj.IsClose == "0" {
									// 重启网关
									rse, err = UDPSend(gatewaySlave, protocol.CMD_GatewayRun, buf)
									if err != nil {
										logger.Error(err)
									}
									logger.Info(rse)
									flag = true
								}
							}
						}
						if !flag {
							return // 端口已经更新了,就关闭这个监控协成
						}
					}
					time.Sleep(10 * time.Second) // 10秒检查一次
				}
			}()
		}
	}()
}

// deployDocker 部署 docker 容器
func deployDocker(project *entity.ProjectDocker, slave string) error {
	taskId := utils.IDMd5()
	portMap := make(map[string]string)
	portList := strings.Split(project.Port, ";")
	for _, v := range portList {
		l := strings.Split(v, ":")
		if len(l) < 2 {
			continue
		}
		portMap[l[0]] = l[1]
	}
	//logger.Info("project.IsGateway = ", utils.Num2Int(project.IsGateway))
	arg := entity.DockerRunArg{
		Image:        project.Image,
		ImageUser:    project.User,
		ImagePass:    project.Password,
		RandomPort:   utils.Num2Int(project.IsGateway),
		PortRelation: portMap,
		Name:         project.Name + "-" + taskId,
		Env:          strings.Split(project.Env, ";"),
		Volume:       strings.Split(project.Volume, ";"),
		TaskId:       taskId,
		IsProject:    true,
		Project:      project.Name,
	}
	//logger.Info("arg.RandomPort = ", arg.RandomPort)
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		logger.Error("部署dokcer失败：", err)
		return err
	}
	_, err = UDPSend(slave, protocol.CMD_DockerRun, buf)
	if err != nil {
		logger.Error("部署dokcer失败：", err)
		return err
	}
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, slave, "部署docker容器项目:"+project.Image)
	logger.Info("docker run 任务启动成功 任务id = ", taskId)
	return nil
}

// TODO ProjectDockerUpdate 更新数据
func ProjectDockerUpdate(c *gin.Context) {
	// 更新数据
	//dockerName := c.Request.PostFormValue("docker_name") // 项目的名称 唯一的
	//dockerNote := c.Request.PostFormValue("docker_note")
	//dockerIsGateway := c.Request.PostFormValue("docker_is_gateway")
	//dockerPort := c.Request.PostFormValue("docker_port")
	//dockerImage := c.Request.PostFormValue("docker_image")
	//dockerUser := c.Request.PostFormValue("docker_user")
	//dockerPassword := c.Request.PostFormValue("docker_password")
	//dockerEnv := c.Request.PostFormValue("docker_env")
	//dockerDuplicate := c.Request.PostFormValue("docker_duplicate")

	// 如果有网关保护则启用灰度发布
	// 更新完成后通知网关更新ips

	// 如果没有网关保护则，先删拉新镜像再删旧容器最后启动新容器

}

// ProjectDockerUpdateImage 更新项目镜像 并重启
func ProjectDockerUpdateImage(c *gin.Context) {
	dockerName := c.Request.PostFormValue("project_name")        // 项目的名称 唯一的
	dockerImage := c.Request.PostFormValue("docker_image")       // 新的镜像
	dockerUser := c.Request.PostFormValue("docker_user")         // 新的镜像拉取账号
	dockerPassword := c.Request.PostFormValue("docker_password") // 新的镜像拉取密码
	if len(dockerName) == 0 || len(dockerImage) == 0 {
		APIOutPut(c, 1, 0, "", "参数错误")
		return
	}
	project, err := new(dao.DaoProjectDocker).Get(dockerName)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	project.Image = dockerImage
	if len(dockerUser) > 0 {
		project.User = dockerUser
	}
	if len(dockerPassword) > 0 {
		project.Password = dockerPassword
	}
	err = new(dao.DaoProjectDocker).Set(dockerName, project)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	// 通知更新， 如果有网关层则灰度发布， 没有则先拉再删最后启动
	if project.IsGateway == "1" {
		// 网关层则灰度发布
		dockerUpdateImageIsGateway(project)
	} else {
		// 先拉再删最后启动
		dockerUpdateImage(project)
	}
	// 修改项目持久化数据
	project.UpdateTime = utils.NowTimeStr()
	err = new(dao.DaoProjectDocker).Set(dockerName, project)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 1, 0, "ok", "ok")
}

// dockerUpdateImageIsGateway 含网关的项目更新，执行灰度发布
func dockerUpdateImageIsGateway(project *entity.ProjectDocker) {
	// 启动一个监视器，监视项目的容器更新情况，更新完成则通知网关更新， 更新失败则撤回更新
	go func() {
		oladContainerData, oladContainerId, err := new(dao.DaoProjectDocker).GetProjectDockerContainer(project.Name)
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Info("旧容器列表 = ", oladContainerId)
		l := utils.Num2Int(project.Duplicate) * 2
		times := 0
		for {
			if times > 180 {
				logger.Info("更新超时..........")
				logger.Info("旧容器列表 = ", oladContainerId)
				// 更新超时， 撤回更新
				return
			}
			_, containerList, _ := new(dao.DaoProjectDocker).GetProjectDockerContainer(project.Name)
			if len(containerList) == l {
				// 更新完成
				logger.Info("更新完成..........")
				logger.Info("旧容器列表 = ", oladContainerId)
				logger.Info("当前容器列表 = ", containerList)
				//newContainerList := utils.SliceDifference(containerList, oladContainerId)
				//logger.Info("新的容器 = ", newContainerList)
				//通知更新网关， 数据是需要去掉旧的
				for _, container := range oladContainerData {
					for k, v := range container.Port {
						arg := &entity.RegisterIpToGatewayArg{
							Key: fmt.Sprintf("%s%s", container.Project, v),
							Ip:  container.Slave + ":" + k,
						}
						logger.Error(arg.Key, arg.Ip)
						buf, err := protocol.DataEncoder(arg)
						if err != nil {
							logger.Error("在网关上注册地址失败")
							return
						}
						rse, err := UDPSend(container.Slave, protocol.CMD_DelRegisterIPGateway, buf)
						if err != nil {
							logger.Error(err)
						}
						logger.Info(rse)
					}
					// 删除旧的，删除旧的容器
					rmArg := &entity.DockerContainerRmArg{
						IsProject:   "0",
						Slave:       container.Slave,
						Project:     "",
						ContainerId: container.ContainerId,
						TaskID:      utils.IDMd5(),
					}
					rmBuf, rmErr := protocol.DataEncoder(rmArg)
					if rmErr != nil {
						logger.Error("删除旧容器失败 err = ", rmErr)
					}
					_, err = UDPSend(container.Slave, protocol.CMD_DockerRm, rmBuf)
					if err != nil {
						logger.Error("删除旧容器失败 err = ", err)
					}
					// 删除旧的数据
					err = new(dao.DaoProjectDocker).DelProjectDockerContainer(container.Project, container.ContainerId)
					if err != nil {
						logger.Error("删除旧容器数据失败 err = ", err)
					}
				}
			}
			time.Sleep(2 * time.Second)
			times++
		}
	}()

	// 部署并运行
	online := protocol.AllUdpClient.GetAllKey()
	n := 0
	l := len(online) - 1
	i := 0
	for i < utils.Num2Int(project.Duplicate) {
		if n >= l {
			n = 0
		}
		// 启动新容器
		deployErr := deployDocker(project, online[n])
		if deployErr == nil {
			i++
		}
		n++
	}
}

// dockerUpdateImage 不会网关的项目更新，先拉再删最后启动
func dockerUpdateImage(project *entity.ProjectDocker) {
	// 获取项目下的所有容器
	oladContainerData, _, err := new(dao.DaoProjectDocker).GetProjectDockerContainer(project.Name)
	if err != nil {
		logger.Error(err)
		return
	}
	// 删除旧的容器
	for _, v := range oladContainerData {
		rmArg := &entity.DockerContainerRmArg{
			IsProject:   "1",
			Slave:       v.Slave,
			Project:     v.Project,
			ContainerId: v.ContainerId,
			TaskID:      v.TaskID,
		}
		buf, err := protocol.DataEncoder(rmArg)
		if err != nil {
			logger.Error(err)
		}
		rse, err := UDPSend(v.Slave, protocol.CMD_DockerRm, buf)
		if err != nil {
			logger.Error(err)
		}
		logger.Info(rse)
	}
	// 启动新容器
	online := protocol.AllUdpClient.GetAllKey()
	n := 0
	l := len(online) - 1
	for i := 0; i < utils.Num2Int(project.Duplicate); i++ {
		if n >= l {
			n = 0
		}
		err = deployDocker(project, online[n])
		if err != nil {
			logger.Error("部署失败")
			continue
		}
		n++
	}
}

// ProjectDockerUpdateDuplicate 更新副本数量
func ProjectDockerUpdateDuplicate(c *gin.Context) {
	projectName := c.Request.PostFormValue("project_name") // 项目的名称 唯一的
	duplicate := c.Request.PostFormValue("duplicate")      // 副本数量
	if utils.Num2Int(duplicate) < 1 {
		APIOutPut(c, 1, 0, "", "副本数量小于1是不允许的，如果你有这个需求，则使用删除所有容器的操作!")
		return
	}
	// 获取项目信息
	project, err := new(dao.DaoProjectDocker).Get(projectName)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	// 获取当前容器数量
	oladContainerData, oladContainerIdList, err := new(dao.DaoProjectDocker).GetProjectDockerContainer(project.Name)
	if err != nil {
		logger.Error(err)
		return
	}
	oladContainerNumber := len(oladContainerIdList)
	// if > 当前容器数量， 超过部分需要部署
	// if < 当前容器数量， 需要删除多余的容器
	if utils.Num2Int(duplicate) > oladContainerNumber {
		difference := utils.Num2Int(duplicate) - oladContainerNumber
		// 启动新容器
		online := protocol.AllUdpClient.GetAllKey()
		n := 0
		l := len(online) - 1
		i := 0
		for i < difference {
			if n >= l {
				n = 0
			}
			deployErr := deployDocker(project, online[n])
			if deployErr == nil {
				i++
			}
			n++
		}
	} else if utils.Num2Int(duplicate) < oladContainerNumber {
		excess := oladContainerNumber - utils.Num2Int(duplicate)
		i := 0
		logger.Info("删除多余的容器 ", excess)
		for _, v := range oladContainerData {
			if i == excess {
				break
			}
			rmArg := &entity.DockerContainerRmArg{
				IsProject:   "1",
				Slave:       v.Slave,
				Project:     v.Project,
				ContainerId: v.ContainerId,
				TaskID:      v.TaskID,
			}
			buf, err := protocol.DataEncoder(rmArg)
			if err != nil {
				APIOutPut(c, 1, 0, "", err.Error())
				return
			}
			rse, err := UDPSend(v.Slave, protocol.CMD_DockerRm, buf)
			if err != nil {
				logger.Error(err)
			}
			logger.Info(rse)
			i++
		}
	}
	APIOutPut(c, 1, 0, "ok", "ok")
}

func ProjectDockerList(c *gin.Context) {
	data, _ := new(dao.DaoProjectDocker).GetALL()
	APIOutPut(c, 1, 0, data, "ok")
}

func ProjectDockerRun(c *gin.Context) {
	// 获取在线的slave数量， 根据副本数分别部署
	// 并且记录部署的数据
	projectName := c.Query("project") // ip
	project, err := new(dao.DaoProjectDocker).Get(projectName)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		err = deployDocker(project, v)
		if err != nil {
			logger.Error(err)
		}
	}
	APIOutPut(c, 1, 0, "部署已经执行，请关注通知", "ok")
}

func ProjectDockerContainer(c *gin.Context) {
	projectName := c.Query("project") // ip
	data, _, err := new(dao.DaoProjectDocker).GetProjectDockerContainer(projectName)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 1, 0, data, "")
	return
}

func ProjectDockerContainerIps(c *gin.Context) {
	projectName := c.Query("project") // ip
	port := c.Query("port")
	rse := make([]string, 0)
	data, _, err := new(dao.DaoProjectDocker).GetProjectDockerContainer(projectName)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	for _, v := range data {
		for k, p := range v.Port {
			if p == port {
				rse = append(rse, v.Slave+":"+k)
			}
		}

	}
	APIOutPut(c, 0, 0, strings.Join(rse, ";"), "")
	return
}
