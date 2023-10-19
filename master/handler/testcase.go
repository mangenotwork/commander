package handler

import (
	"time"

	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/gin-gonic/gin"
)

func TestNotice(c *gin.Context) {
	BroadcastNotice([]byte("发送一个测试通知 : " + utils.Int642Str(time.Now().Unix())))
	APIOutPut(c, 0, 0, "", "ok")
}

func TestDockerState(c *gin.Context) {
	slave := c.Query("slave") // ip
	containerId := c.Query("container")
	logger.Info("containerId = ", containerId)
	UDPSendOutHttp(c, slave, protocol.CMD_DockerStateS, []byte(containerId))
}
