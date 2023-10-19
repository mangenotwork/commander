package udp_server

import (
	"time"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"
	"gitee.com/mangenotework/commander/master/handler"
)

func CreateHttpProxy(ctx *HandlerCtx) {
	rse := &entity.ProxyHttpCreateRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("解析slave返回的包错误: ", err.Error())
		return
	}
	logger.Info("CreateHttpProxy data = ", rse)
	// 保存数据
	err = new(dao.DaoHttpsProxy).Set(rse.Name, &entity.HttpsProxy{
		Name:    rse.Name,
		Slave:   rse.Slave,
		Port:    rse.Port,
		Create:  utils.NowTimeStr(),
		IsClose: "0",
		Note:    rse.Note,
	})
	if err != nil {
		logger.Error("保存数据错误，err = ", err)
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId + "]: 启动http/s代理 " +
		rse.Slave + ":" + rse.Port))
}

func UpdateHttpProxy(ctx *HandlerCtx) {
	rse := &entity.ProxyHttpUpdateRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("解析slave返回的包错误: ", err.Error())
		return
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId + "]: 更新 http/s代理 " +
		rse.Name + "  " + rse.Rse))
}

func CreateSocket5Proxy(ctx *HandlerCtx) {
	rse := &entity.Socket5ProxyCreateRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("解析slave返回的包错误: ", err.Error())
		return
	}
	// 保存数据
	err = new(dao.DaoSocket5Proxy).Set(rse.Name, &entity.Socket5Proxy{
		Name:    rse.Name,
		Slave:   rse.Slave,
		Port:    rse.Port,
		Create:  utils.NowTimeStr(),
		IsClose: "0",
		Note:    rse.Note,
	})
	if err != nil {
		logger.Error("保存数据错误，err = ", err)
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId + "]: 启动socket5代理 " +
		rse.Slave + ":" + rse.Port))
}

func UpdateSocket5Proxy(ctx *HandlerCtx) {
	rse := &entity.ProxyHttpUpdateRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("解析slave返回的包错误: ", err.Error())
		return
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId + "]: 更新 socket5代理 " +
		rse.Name + "  " + rse.Rse))
}

func CreateTCPForward(ctx *HandlerCtx) {
	rse := &entity.TCPForwardCreateRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("解析slave返回的包错误: ", err.Error())
		return
	}
	err = new(dao.DaoTCPForward).Set(rse.Name, &entity.TCPForward{
		Name:         rse.Name,
		Slave:        rse.Slave,
		Port:         rse.Port,
		Create:       utils.NowTimeStr(),
		Note:         rse.Note,
		ForwardTable: rse.ForwardTable, // 转发表
		IsClose:      "0",              // 是否关闭  0 否  1 是
		IsDel:        "0",              // 是否删除  0 否 1 是
	})
	if err != nil {
		logger.Error("解析slave返回的包错误: ", err.Error())
		return
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId + "]: 创建TCP 转发 " +
		rse.Name + "  " + rse.Rse))
}

func UpdateTCPForward(ctx *HandlerCtx) {
	rse := &entity.TCPForwardUpdateRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("解析slave返回的包错误: ", err.Error())
		return
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId + "]: 更新 TCP转发 " +
		rse.Name + "  " + rse.Rse))
}
