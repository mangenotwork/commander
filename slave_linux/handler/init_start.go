package handler

import (
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/slve_linux/dao"
	"gitee.com/mangenotework/commander/slve_linux/gateway"
)

func InitStart(){
	// 启动持久化的网关
	OpenGateway()

	gateway.TimingPersistenceIps()
}

func OpenGateway(){
	gatewayList, _ := new(dao.DaoGateway).GetALL()
	logger.Info("启动持久化的网关...... gatewayList = ", gatewayList)
	for _, v := range gatewayList {
		if v.IsClose == "0" {
			arg := &entity.GatewayArg{
				Ports : v.Ports,
				ProjectName : v.ProjectName,
				LVS : v.LVS,
				LVSModel : v.LVSModel,
			}
			gateway.RunGateway(arg)
		}
	}
}
