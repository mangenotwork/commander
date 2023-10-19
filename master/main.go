package main

import (
	"net/http"
	_ "net/http/pprof"

	"gitee.com/mangenotework/commander/common/check"
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/master/cron"
	"gitee.com/mangenotework/commander/master/dao"
	"gitee.com/mangenotework/commander/master/http_server"
	"gitee.com/mangenotework/commander/master/udp_server"
)

func main(){
	conf.MasterInitConf()
	check.MasterInitPath()
	dao.DBInit()
	logger.Info("启动服务......")

	// 启动 Master UDP 服务
	udp_server.RunUDPServer()

	// 启动 Web客户端
	http_server.RunHttpServer()

	// 启动 http/pprof
	go func() {
		_=http.ListenAndServe("127.0.0.1:6060", nil)
	}()

	// 启动定时任务
	go func() {
		cron.ClearPerformance()
	}()

	select {}

}