package http_server

import (
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/master/routers"

	"github.com/gin-gonic/gin"
)

func RunHttpServer() {
	go func() {
		gin.SetMode(gin.DebugMode)
		s := routers.Routers()
		s.Run(":" + conf.MasterConf.HttpServer.Prod)
	}()
}
