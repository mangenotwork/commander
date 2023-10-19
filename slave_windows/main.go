package main

import (
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/slave_windows/handler"
	"log"
)

func main(){
	conf.SlaveInitConf()
	log.Println("启动服务......")
	protocol.UDPClient(handler.InitHandler(), handler.InitFunc, handler.ErrFunc)
	for{}
}