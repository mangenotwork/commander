package handler

import (
	"gitee.com/mangenotework/commander/common/protocol"
	"github.com/gin-gonic/gin"
)

// DeployedInstallDocker 安装docker
func DeployedInstallDocker(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_InstallDocker, []byte(""))
}

// DeployedRemoveDocker 卸载docker
func DeployedRemoveDocker(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_RemoveDocker, []byte(""))
}

// DeployedInstallNginx 安装nginx
func DeployedInstallNginx(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_InstallNginx, []byte(""))
}

// DeployedRemoveNginx 卸载nginx
func DeployedRemoveNginx(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_RemoveNginx, []byte(""))
}
