package handler

import (
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"
	"github.com/gin-gonic/gin"
)

// NginxInfo nginx 信息
func NginxInfo(c *gin.Context) {
	slave := c.Query("slave")
	taskId := utils.IDMd5()
	arg := entity.NginxInfoArg{
		TaskId: taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_NginxInfo, buf)
}

// NginxStart 启动nginx
func NginxStart(c *gin.Context) {
	slave := c.Query("slave") // ip
	task := utils.IDMd5()
	// 任务记录
	new(dao.DaoTask).SetDefaultCreate(slave, task, "启动nginx")
	// 下发指令
	arg := &entity.NginxStartArg{
		TaskId: task,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_NginxStart, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "已下发")
}

// NginxReload 重启nginx -s reload
func NginxReload(c *gin.Context) {
	slave := c.Query("slave") // ip
	task := utils.IDMd5()
	// 任务记录
	new(dao.DaoTask).SetDefaultCreate(slave, task, "重启nginx")
	// 下发指令
	arg := &entity.NginxReloadArg{
		TaskId: task,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_NginxReload, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "已下发")
}

// NginxQuit 停止nginx nginx -s quit
func NginxQuit(c *gin.Context) {
	slave := c.Query("slave") // ip
	task := utils.IDMd5()
	// 任务记录
	new(dao.DaoTask).SetDefaultCreate(slave, task, "停止nginx")
	// 下发指令
	arg := &entity.NginxQuitArg{
		TaskId: task,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_NginxQuit, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "已下发")
}

// NginxStop  强制停止nginx nginx -s stop
func NginxStop(c *gin.Context) {
	slave := c.Query("slave") // ip
	task := utils.IDMd5()
	// 任务记录
	new(dao.DaoTask).SetDefaultCreate(slave, task, "强制停止nginx")
	// 下发指令
	arg := &entity.NginxStopArg{
		TaskId: task,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_NginxStop, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "已下发")
}

// NginxCheckConf 检查nginx配置
func NginxCheckConf(c *gin.Context) {
	slave := c.Query("slave") // ip
	task := utils.IDMd5()
	arg := &entity.NginxCheckConfArg{
		TaskId: task,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_NginxCheckConf, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "已下发")
}

type NginxConfUpdateParam struct {
	Slave    string `json:"slave"`
	ConfPath string `json:"path"`
	ConfData string `json:"data"`
}

// NginxConfUpdate 更新nginx conf文件
func NginxConfUpdate(c *gin.Context) {
	param := &NginxConfUpdateParam{}
	err := c.BindJSON(param)
	if err != nil {
		logger.Infof("json err = ", err)
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	taskId := utils.IDMd5()
	arg := entity.NginxConfUpdateArg{
		TaskId: taskId,
		Path:   param.ConfPath,
		Data:   param.ConfData,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, param.Slave, protocol.CMD_NginxConfUpdate, buf)
}
