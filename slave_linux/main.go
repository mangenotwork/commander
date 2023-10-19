package main

import (
	"gitee.com/mangenotework/commander/common/check"
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/slve_linux/dao"
	"gitee.com/mangenotework/commander/slve_linux/fileser"
	"gitee.com/mangenotework/commander/slve_linux/handler"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main(){
	conf.SlaveInitConf()

	check.SlaveInitPath()

	dao.DBInit()
	log.Println("启动服务......")

	// 初始化启动
	handler.InitStart()

	go protocol.UDPClient(handler.InitHandler(), handler.InitFunc, handler.ErrFunc)

	// http/pprof
	go func() {
		http.ListenAndServe("127.0.0.1:7070", nil)
	}()

	// file server
	go func() {
		fileser.Run()
	}()

	select {

	}
}