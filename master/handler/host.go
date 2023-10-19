package handler

import (
	"fmt"
	"strings"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

func UDPSend(slave string, cmd protocol.CommandCode, data []byte) (string, error) {
	udpC, ok := protocol.AllUdpClient.RetryGet(slave)
	if !ok {
		logger.Infof(slave, "  离线 ok is false")
	}
	logger.Infof("udpC = ", udpC)
	if udpC == nil {
		logger.Info(slave, "  离线  udpC = nil ")
		return "", fmt.Errorf("udp is null")
	}
	requst := utils.IDMd5()
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

func UDPSendOutHttp(c *gin.Context, slave string, cmd protocol.CommandCode, data []byte) {
	logger.Infof("UDPSendOutHttp CommandCode = ", cmd)
	udpC, ok := protocol.AllUdpClient.RetryGet(slave)
	if udpC == nil || !ok {
		c.JSON(200, slave+" 离线  udpC = nil ")
		return
	}
	requst := utils.IDMd5() // 6854823418404110336
	logger.Info("requst = ", requst)
	logger.Info("CommandCode = ", cmd.Chinese())
	packate, err := protocol.Packet(cmd, requst, data)
	if err != nil {
		logger.Error(err)
	}
	//logger.Infof("发送数据: ", packate)
	logger.Info(udpC.Conn, udpC.IP)
	protocol.UDPSend(udpC.Conn, packate)
	protocol.Set(requst)
	tx, err := protocol.Get(requst)
	if err != nil {
		c.JSON(200, err)
	}
	defer protocol.Close(requst)
	select {
	case err = <-tx.Err:
		c.JSON(200, err)
	case rse := <-tx.Data:
		//logger.Infof("protocol.Get(requst) data = ", rse)
		APIOutPut(c, 0, 1, rse, "")
	}
}

func SlaveInfo(c *gin.Context) {
	slave := c.Query("slave") // ip
	udpC, ok := protocol.AllUdpClient.RetryGet(slave)
	if !ok {
		logger.Infof(slave, "  离线 ok is false")
	}
	//logger.Infof("udpC = ", udpC)
	if udpC == nil {
		logger.Infof(slave, "  离线  udpC = nil ")
		return
	}
	requst := utils.IDMd5() // 6854823418404110336
	//logger.Infof("requst = ", requst)
	packate, err := protocol.Packet(protocol.CMD_ReportHostInfo, requst, []byte(""))
	if err != nil {
		APIOutPut(c, 0, 1, "", err.Error())
		return
	}
	//logger.Infof("发送数据: ", packate)
	//logger.Info(udpC.Conn, udpC.IP)
	protocol.UDPSend(udpC.Conn, packate)
	protocol.Set(requst)
	tx, err := protocol.Get(requst)
	if err != nil {
		APIOutPut(c, 0, 1, "", err.Error())
		return
	}
	defer protocol.Close(requst)
	select {
	case err = <-tx.Err:
		APIOutPut(c, 0, 1, "", err.Error())
	case data := <-tx.Data:
		logger.Infof("protocol.Get(requst) data = ", data)
		APIOutPut(c, 0, 1, data, "")
	}
}

// HaveDocker 是否存在docker
func HaveDocker(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_HaveDocker, []byte(""))
}

func DockerInfo(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_DockerInfo, []byte(""))
}

func DockerPS(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_DockerPS, []byte(""))
}

func DockerImages(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_DockerImages, []byte(""))
}

// ImagesDeployParam 镜像部署参数
type ImagesDeployParam struct {
	Slave         string `json:"slave"`
	ImageId       string `json:"image_id"`
	ContainerEnv  string `json:"container_env"`
	ContainerName string `json:"container_name"`
	ContainerPort string `json:"container_port"`
	IsAlways      bool   `json:"is_always"`
}

// DockerImagesDeploy 运行指定镜像
func DockerImagesDeploy(c *gin.Context) {
	// 业务逻辑
	// 1. 检查镜像是否存在
	// 2. 如果存在则执行
	param := &ImagesDeployParam{}
	err := c.BindJSON(param)
	if err != nil {
		logger.Infof("json err = ", err)
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(param.Slave, taskId, "运行docker:"+param.ImageId)
	// 下发任务
	portMap := make(map[string]string)
	portList := strings.Split(param.ContainerPort, ";")
	for _, v := range portList {
		logger.Info(v)
		l := strings.Split(v, ":")
		if len(l) < 2 {
			continue
		}
		portMap[l[0]] = l[1]
	}
	arg := entity.DockerImageRunArg{
		ImageId:      param.ImageId,
		PortRelation: portMap,
		Name:         param.ContainerName,
		Env:          strings.Split(param.ContainerEnv, ";"),
		TaskId:       taskId,
		IsAlways:     param.IsAlways,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(param.Slave, protocol.CMD_DockerImageRun, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "docker run 任务启动成功 任务id = "+taskId)
	return
}

// DockerStop docker stop
func DockerStop(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerId := c.Query("container")
	task := utils.IDMd5()
	// 任务记录
	new(dao.DaoTask).SetDefaultCreate(slave, task, "停止容器:"+containerId)
	// 下发指令
	arg := &entity.DockerStopArg{
		TaskId:      task,
		ContainerId: containerId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_DockerStop, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "已下发")
}

// DockerRm docker rm
func DockerRm(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerId := c.Query("container")
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "删除容器："+containerId)
	// 下发指令
	rmArg := &entity.DockerContainerRmArg{
		IsProject:   "0",
		Slave:       slave,
		Project:     "",
		ContainerId: containerId,
		TaskID:      taskId,
	}
	buf, err := protocol.DataEncoder(rmArg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_DockerRm, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "已下发")
}

func DockerRmi(c *gin.Context) {
	slave := c.Query("slave") // ip
	imageId := c.Query("image")
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "删除镜像:"+imageId)
	// 下发通知
	arg := entity.DockerRmiArg{
		TaskId:  taskId,
		ImageId: imageId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_DockerRmi, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "已下发")
}

// DockerPull docker pull
func DockerPull(c *gin.Context) {
	slave := c.Query("slave") // ip
	name := c.Query("name")
	pass := c.Query("pass")
	image := c.Query("image")
	logger.Info(image, name, pass)
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "拉取镜像:"+name)
	// 下发任务
	pull := entity.DockerPullArg{
		Name:   name,
		Pass:   pass,
		Image:  image,
		TaskId: taskId,
	}
	logger.Info("pull = ", pull)
	buf, err := protocol.DataEncoder(pull)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_DockerPull, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "docker pull 任务启动成功 任务id = "+taskId)
	return
}

// GetTaskState 获取任务状态
func GetTaskState(c *gin.Context) {
	taskId := c.Query("task")
	task, err := new(dao.DaoTask).Get(taskId)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 1, task, "")
	return
}

// DockerRunParam  docker run param
type DockerRunParam struct {
	Image           string `json:"image"`
	ImageUser       string `json:"image_user"`
	ImagePass       string `json:"image_pass"`
	ContainerEnv    string `json:"container_env"`
	ContainerName   string `json:"container_name"`
	ContainerPort   string `json:"container_port"`
	ContainerVolume string `json:"container_volume"`
	IsAlways        bool   `json:"is_always"`
}

func DockerRun(c *gin.Context) {
	slave := c.Query("slave") // ip
	param := &DockerRunParam{}
	err := c.BindJSON(param)
	if err != nil {
		logger.Infof("json err = ", err)
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "运行docker:"+param.Image)
	// 下发任务
	portMap := make(map[string]string)
	portList := strings.Split(param.ContainerPort, ";")
	for _, v := range portList {
		l := strings.Split(v, ":")
		portMap[l[0]] = l[1]
	}
	arg := entity.DockerRunArg{
		Image:        param.Image,
		ImageUser:    param.ImageUser,
		ImagePass:    param.ImagePass,
		PortRelation: portMap,
		Name:         param.ContainerName,
		Env:          strings.Split(param.ContainerEnv, ";"),
		Volume:       strings.Split(param.ContainerVolume, ";"),
		TaskId:       taskId,
		IsAlways:     param.IsAlways,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_DockerRun, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "docker run 任务启动成功 任务id = "+taskId)
	return
}

func DockerTop(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerId := c.Query("container")
	UDPSendOutHttp(c, slave, protocol.CMD_ContainerTop, []byte(containerId))
}

func DockerReName(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerId := c.Query("container")
	newName := c.Query("name")
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "修改容器名称:"+containerId)
	// 下发任务
	arg := entity.ContainerReNameArg{
		ContainerId: containerId,
		NewName:     newName,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_ContainerRename, buf)
}

func DockerRestart(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerId := c.Query("container")
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "重启容器："+containerId)
	// 下发任务
	arg := &entity.DockerRestartArg{
		TaskId:      taskId,
		ContainerId: containerId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_ContainerRestart, buf)
}

func DockerPause(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerId := c.Query("container")
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "暂停容器:"+containerId)
	// 下发任务
	arg := &entity.DockerPauseArg{
		TaskId:      taskId,
		ContainerId: containerId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_ContainerPause, buf)
}

func DockerStates(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerIds := c.Query("containers")
	UDPSendOutHttp(c, slave, protocol.CMD_DockerStateS, []byte(containerIds))
}

func DockerState(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerId := c.Query("container")
	UDPSendOutHttp(c, slave, protocol.CMD_DockerStateS, []byte(containerId))
}
