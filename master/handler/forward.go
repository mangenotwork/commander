package handler

import (
	"strings"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

// ProxyHttpCreate 创建http/s代理
func ProxyHttpCreate(c *gin.Context) {
	name := c.Request.PostFormValue("name")   // 代理名称 用于表示唯一性
	note := c.Request.PostFormValue("note")   // 代理备注
	slave := c.Request.PostFormValue("slave") // 部署在哪个主机上
	port := c.Request.PostFormValue("port")   // 代理的端口
	taskId := utils.IDMd5()
	arg := &entity.ProxyHttpCreateArg{
		Name:   name,
		Slave:  slave,
		Port:   port,
		Note:   note,
		TaskId: taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_CreateHttpProxy, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	APIOutPut(c, 1, 0, "", "已下发任务:"+taskId)
}

// ProxyHttpList 获取http/s代理列表
func ProxyHttpList(c *gin.Context) {
	data, _ := new(dao.DaoHttpsProxy).GetALL()
	APIOutPut(c, 0, 0, data, "")
}

// ProxyHttpStop 获取http/s代理暂停
func ProxyHttpStop(c *gin.Context) {
	updateHttpProxy(c, "IsClose", "1")
}

// ProxyHttpContinue 获取http/s代理继续
func ProxyHttpContinue(c *gin.Context) {
	updateHttpProxy(c, "IsClose", "0")
}

// ProxyHttpRemove 删除http/s代理
func ProxyHttpRemove(c *gin.Context) {
	updateHttpProxy(c, "IsDel", "1")
}

// updateHttpProxy
func updateHttpProxy(c *gin.Context, updateType, vlaue string) {
	name := c.Query("name")
	proxyObj, err := new(dao.DaoHttpsProxy).Get(name)
	if err != nil {
		APIOutPut(c, 1, 0, "", "暂停失败，未找到该代理 : "+err.Error())
		return
	}
	if updateType == "IsClose" {
		proxyObj.IsClose = vlaue
	}
	if updateType == "IsDel" {
		proxyObj.IsDel = vlaue
	}
	err = new(dao.DaoHttpsProxy).Set(name, proxyObj)
	if err != nil {
		APIOutPut(c, 1, 0, "", "暂停失败 : "+err.Error())
		return
	}
	// 下行通知
	taskId := utils.IDMd5()
	arg := &entity.ProxyHttpUpdateArg{
		Name:       name,
		UpdateType: updateType,
		Vlaue:      vlaue,
		TaskId:     taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	_, err = UDPSend(proxyObj.Slave, protocol.CMD_UpdateHttpProxy, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	APIOutPut(c, 1, 0, "", "已下发任务")
}

// ProxySocket5Create 创建socket5代理
func ProxySocket5Create(c *gin.Context) {
	name := c.Request.PostFormValue("name")   // 代理名称 用于表示唯一性
	note := c.Request.PostFormValue("note")   // 代理备注
	slave := c.Request.PostFormValue("slave") // 部署在哪个主机上
	port := c.Request.PostFormValue("port")   // 代理的端口
	// 下行通知
	taskId := utils.IDMd5()
	arg := &entity.Socket5ProxyCreateArg{
		Name:   name,
		Slave:  slave,
		Port:   port,
		Note:   note,
		TaskId: taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_CreateSocket5Proxy, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}

	APIOutPut(c, 1, 0, "", "已下发任务:"+taskId)
}

// ProxySocket5List 获取socket5代理列表
func ProxySocket5List(c *gin.Context) {
	data, _ := new(dao.DaoSocket5Proxy).GetALL()
	APIOutPut(c, 0, 0, data, "")
}

// ProxySocket5Stop 获取socket5代理暂停
func ProxySocket5Stop(c *gin.Context) {
	updateSocket5Proxy(c, "IsClose", "1")
}

// ProxySocket5Continue 获取http/s代理继续
func ProxySocket5Continue(c *gin.Context) {
	updateSocket5Proxy(c, "IsClose", "0")
}

// ProxySocket5Remove 删除socket5代理
func ProxySocket5Remove(c *gin.Context) {
	updateHttpProxy(c, "IsDel", "1")
}

// updateSocket5Proxy
func updateSocket5Proxy(c *gin.Context, updateType, vlaue string) {
	name := c.Query("name")
	proxyObj, err := new(dao.DaoSocket5Proxy).Get(name)
	if err != nil {
		APIOutPut(c, 1, 0, "", "暂停失败，未找到该代理 : "+err.Error())
		return
	}
	if updateType == "IsClose" {
		proxyObj.IsClose = vlaue
	}
	if updateType == "IsDel" {
		proxyObj.IsDel = vlaue
	}
	err = new(dao.DaoSocket5Proxy).Set(name, proxyObj)
	if err != nil {
		APIOutPut(c, 1, 0, "", "暂停失败 : "+err.Error())
		return
	}
	// 下行通知
	taskId := utils.IDMd5()
	arg := &entity.ProxySocket5UpdateArg{
		Name:       name,
		UpdateType: updateType,
		Vlaue:      vlaue,
		TaskId:     taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	_, err = UDPSend(proxyObj.Slave, protocol.CMD_UpdateSocket5Proxy, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	APIOutPut(c, 1, 0, "", "已下发任务")
}

// ForwardTCPCreate 创建TCP转发
func ForwardTCPCreate(c *gin.Context) {
	name := c.Request.PostFormValue("name")                  // 代理名称 用于表示唯一性
	note := c.Request.PostFormValue("note")                  // 代理备注
	slave := c.Request.PostFormValue("slave")                // 部署在哪个主机上
	port := c.Request.PostFormValue("port")                  // 代理的端口
	forwardTable := c.Request.PostFormValue("forward_table") // 转发表
	forwardTableList := strings.Split(forwardTable, ";")
	forwardTableData := make([]string, 0)
	for _, v := range forwardTableList {
		if v != "" {
			forwardTableData = append(forwardTableData, v)
		}
	}
	taskId := utils.IDMd5()
	arg := &entity.TCPForwardCreateArg{
		Name:         name,
		Slave:        slave,
		Port:         port,
		Note:         note,
		TaskId:       taskId,
		ForwardTable: forwardTableData,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	_, err = UDPSend(slave, protocol.CMD_CreateTCPForward, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	APIOutPut(c, 1, 0, "", "已下发任务:"+taskId)
}

// ForwardTCPList 获取TCP转发列表
func ForwardTCPList(c *gin.Context) {
	data, _ := new(dao.DaoTCPForward).GetALL()
	APIOutPut(c, 0, 0, data, "")
}

// ForwardTCPCutoff 切断TCP转发
func ForwardTCPCutoff(c *gin.Context) {
	updateTCPForwardProxy(c, "IsClose", "1")
}

// ForwardTCPRenew 恢复TCP转发
func ForwardTCPRenew(c *gin.Context) {
	updateTCPForwardProxy(c, "IsClose", "0")
}

// ForwardTCPRemove 删除TCP转发
func ForwardTCPRemove(c *gin.Context) {
	updateTCPForwardProxy(c, "IsDel", "1")
}

// updateTCPForwardProxy
func updateTCPForwardProxy(c *gin.Context, updateType, vlaue string) {
	name := c.Query("name")
	forwardObj, err := new(dao.DaoTCPForward).Get(name)
	if err != nil {
		APIOutPut(c, 1, 0, "", "暂停失败，未找到该代理 : "+err.Error())
		return
	}
	if updateType == "IsClose" {
		forwardObj.IsClose = vlaue
	}
	if updateType == "IsDel" {
		forwardObj.IsDel = vlaue
	}
	if updateType == "ForwardTable" {
		forwardTableList := strings.Split(vlaue, ";")
		forwardTableData := make([]string, 0)
		for _, v := range forwardTableList {
			if v != "" {
				forwardTableData = append(forwardTableData, v)
			}
		}
		forwardObj.ForwardTable = forwardTableData
	}
	err = new(dao.DaoTCPForward).Set(name, forwardObj)
	if err != nil {
		APIOutPut(c, 1, 0, "", "暂停失败 : "+err.Error())
		return
	}
	// 下行通知
	taskId := utils.IDMd5()
	arg := &entity.TCPForwardUpdateArg{
		Name:       name,
		UpdateType: updateType,
		Vlaue:      vlaue,
		TaskId:     taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	_, err = UDPSend(forwardObj.Slave, protocol.CMD_UpdateTCPForward, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", "创建失败:"+err.Error())
		return
	}
	APIOutPut(c, 1, 0, "", "已下发任务")
}

// ForwardTCPSwitch 切换TCP转发目标
func ForwardTCPSwitch(c *gin.Context) {
	table := c.Query("table")
	logger.Info("table = ", table)
	updateTCPForwardProxy(c, "ForwardTable", table)
}

// TODO ForwardUDPCreate 创建UDP转发
func ForwardUDPCreate(c *gin.Context) {

}

// TODO ForwardUDPList 获取UDP转发列表
func ForwardUDPList(c *gin.Context) {

}

// TODO ForwardUDPRemove 删除UDP转发
func ForwardUDPRemove(c *gin.Context) {

}

// TODO ForwardUDPSwitch 切换UDP转发目标
func ForwardUDPSwitch(c *gin.Context) {

}

// TODO ForwardUDPCutoff 切断UDP转发
func ForwardUDPCutoff(c *gin.Context) {

}

// TODO ForwardUDPRenew 恢复UDP转发
func ForwardUDPRenew(c *gin.Context) {

}

// TODO ProxySSHCreate 创建ssh代理
func ProxySSHCreate(c *gin.Context) {

}

// TODO ProxySSHList 获取ssh代理列表
func ProxySSHList(c *gin.Context) {

}

// TODO ProxySSHRemove 删除ssh代理
func ProxySSHRemove(c *gin.Context) {

}
