package handler

import (
	"strings"

	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

// GatewayRun  运行网关
func GatewayRun(c *gin.Context) {
	slave := c.Query("slave")
	port := c.Query("port")
	project := c.Query("project")
	deployGateway(project, port, slave)
	APIOutPut(c, 1, 0, "以下发部署命令", "ok")
}

// GatewayNew 新建网关
func GatewayNew(c *gin.Context) {
	gatewayName := c.Request.PostFormValue("gateway_name") //网关服务名称
	gatewayPort := c.Request.PostFormValue("gateway_port") //网关服务入网端口
	//forwardTable := c.Request.PostFormValue("forward_table") //默认转发表
	gatewaySlave := c.Request.PostFormValue("gateway_slave") //部署在哪个服务器

	//logger.Info("gateway_name = ", gatewayName)
	//logger.Info("gateway_port = ", gatewayPort)
	////logger.Info("forward_table = ", forwardTable)
	//logger.Info("gateway_slave = ", gatewaySlave)

	deployGateway(gatewayName, gatewayPort, gatewaySlave)
}

func GatewayList(c *gin.Context) {
	data, _ := new(dao.DaoGateway).GetALL()
	APIOutPut(c, 1, 0, data, "ok")
}

// GatewayDelete 删除一个网关
func GatewayDelete(c *gin.Context) {
	slave := c.Query("slave")
	project := c.Query("project")
	gatewayObj, err := new(dao.DaoGateway).Get(project)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	gatewayObj.IsClose = "1"
	err = new(dao.DaoGateway).Set(project, gatewayObj)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	buf, err := protocol.DataEncoder(project)
	if err != nil {
		logger.Error("启动网关失败")
		return
	}
	rse, err := UDPSend(slave, protocol.CMD_GatewayDel, buf)
	if err != nil {
		logger.Error(err)
	}
	logger.Info(rse)
}

// TODO GatewayUpdatePort 修改网关端口映射
func GatewayUpdatePort(c *gin.Context) {
	projectName := c.Request.PostFormValue("project_name")
	gatewaySlave := c.Request.PostFormValue("gateway_slave")
	gatewayPort := c.Request.PostFormValue("gateway_port")
	// 修改端口
	gatewayObj, err := new(dao.DaoGateway).Get(projectName)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	gatewayObj.Ports = strings.Split(gatewayPort, ";")
	err = new(dao.DaoGateway).Set(projectName, gatewayObj)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	// 关闭旧网关
	err = new(dao.DaoGateway).Set(projectName, gatewayObj)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	buf, err := protocol.DataEncoder(projectName)
	if err != nil {
		logger.Error("启动网关失败")
		return
	}
	rse, err := UDPSend(gatewaySlave, protocol.CMD_GatewayDel, buf)
	if err != nil {
		logger.Error(err)
	}
	logger.Info(rse)

	// 启动新网关
	deployGateway(projectName, gatewayPort, gatewaySlave)
	APIOutPut(c, 1, 0, "已下发部署命令", "ok")
}
