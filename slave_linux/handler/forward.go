package handler

import (
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/slve_linux/dao"
	"gitee.com/mangenotework/commander/slve_linux/proxy"
	"strings"
)

// CreateHttpProxy 部署并启动一个http/s代理
func CreateHttpProxy(ctx *protocol.HandlerCtx) {
	arg := &entity.ProxyHttpCreateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	logger.Info("部署并启动一个http/s代理 arg  = ", arg)

	rse := &entity.ProxyHttpCreateRse{
		Name:   arg.Name,
		Slave:  arg.Slave,
		Port:   arg.Port,
		Note:   arg.Note,
		TaskId: arg.TaskId,
	}
	//启动代理
	proxy.HttpProxyServer(arg.Port, arg.Name)

	// 持久化数据
	err = new(dao.DaoHttpsProxy).Set(arg.Name, &entity.HttpsProxy{
		Name:    arg.Name,
		Slave:   arg.Slave,
		Port:    arg.Port,
		Create:  utils.NowTimeStr(),
		IsClose: "0",
		Note:    arg.Note,
		IsDel:   "0",
	})
	if err != nil {
		logger.Error(err)
	}

	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}

}

// UpdateHttpProxy 修改http/s代理, 常用功能是 暂停，继续，删除
func UpdateHttpProxy(ctx *protocol.HandlerCtx) {
	arg := &entity.ProxyHttpUpdateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}

	rse := &entity.ProxyHttpUpdateRse{
		Name:   arg.Name,
		TaskId: arg.TaskId,
	}

	proxyObj, err := new(dao.DaoHttpsProxy).Get(arg.Name)
	if err != nil {
		rse.Rse = err.Error()
		goto R
	}

	if arg.UpdateType == "IsClose" {
		proxyObj.IsClose = arg.Vlaue
	}
	// 删除
	if arg.UpdateType == "IsDel" {
		proxyObj.IsDel = arg.Vlaue
	}
	err = new(dao.DaoHttpsProxy).Set(arg.Name, proxyObj)
	if err != nil {
		rse.Rse = err.Error()
	}
	if rse.Rse == "" {
		rse.Rse = "成功"
	}

R:
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// CreateSocket5Proxy 创建socket5代理
func CreateSocket5Proxy(ctx *protocol.HandlerCtx) {
	arg := &entity.Socket5ProxyCreateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}
	logger.Info("部署并启动一个http/s代理 arg  = ", arg)

	rse := &entity.Socket5ProxyCreateRse{
		Name:   arg.Name,
		Slave:  arg.Slave,
		Port:   arg.Port,
		Note:   arg.Note,
		TaskId: arg.TaskId,
	}
	//启动代理
	proxy.SocketProxy(arg.Port, arg.Name)

	// 持久化数据
	err = new(dao.DaoSocket5Proxy).Set(arg.Name, &entity.Socket5Proxy{
		Name:    arg.Name,
		Slave:   arg.Slave,
		Port:    arg.Port,
		Create:  utils.NowTimeStr(),
		IsClose: "0",
		Note:    arg.Note,
		IsDel:   "0",
	})
	if err != nil {
		logger.Error(err)
	}

	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// UpdateSocket5Proxy 修改socket5代理 常用功能是 暂停，继续，删除
func UpdateSocket5Proxy(ctx *protocol.HandlerCtx) {
	arg := &entity.ProxySocket5UpdateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}

	rse := &entity.ProxySocket5UpdateRse{
		Name:   arg.Name,
		TaskId: arg.TaskId,
	}

	proxyObj, err := new(dao.DaoSocket5Proxy).Get(arg.Name)
	if err != nil {
		rse.Rse = err.Error()
		goto R
	}

	if arg.UpdateType == "IsClose" {
		proxyObj.IsClose = arg.Vlaue
	}
	// 删除
	if arg.UpdateType == "IsDel" {
		proxyObj.IsDel = arg.Vlaue
	}
	err = new(dao.DaoSocket5Proxy).Set(arg.Name, proxyObj)
	if err != nil {
		rse.Rse = err.Error()
	}
	if rse.Rse == "" {
		rse.Rse = "成功"
	}

R:
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// CreateTCPForward 创建tcp转发
func CreateTCPForward(ctx *protocol.HandlerCtx) {
	arg := &entity.TCPForwardCreateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}

	rse := &entity.TCPForwardCreateRse{
		Name:         arg.Name,
		Slave:        arg.Slave,
		Port:         arg.Port,
		Note:         arg.Note,
		ForwardTable: arg.ForwardTable,
		TaskId:       arg.TaskId,
	}

	err = new(dao.DaoTCPForward).Set(arg.Name, &entity.TCPForward{
		Name:         arg.Name,
		Slave:        arg.Slave,
		Port:         arg.Port,
		Create:       utils.NowTimeStr(),
		Note:         arg.Note,
		ForwardTable: arg.ForwardTable, // 转发表
		IsClose:      "0",              // 是否关闭  0 否  1 是
		IsDel:        "0",              // 是否删除  0 否 1 是
	})

	if err != nil {
		rse.Error = err.Error()
		rse.Rse = "失败"
		goto R
	}

	go proxy.TCPForward(arg.Name, arg.Port)
	rse.Rse = "成功"

R:
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}

// UpdateTCPForward 修改tcp转发 常用功能是 暂停，继续，删除
func UpdateTCPForward(ctx *protocol.HandlerCtx) {
	arg := &entity.TCPForwardUpdateArg{}
	err := protocol.DataDecoder(ctx.Stream.Data, &arg)
	if err != nil {
		logger.Error("接收参数失败 err = ", err)
		return
	}

	rse := &entity.TCPForwardUpdateRse{
		Name:   arg.Name,
		TaskId: arg.TaskId,
	}

	forwardObj, err := new(dao.DaoTCPForward).Get(arg.Name)
	if err != nil {
		rse.Rse = err.Error()
		goto R
	}
	if arg.UpdateType == "IsClose" {
		forwardObj.IsClose = arg.Vlaue
	}
	// 删除
	if arg.UpdateType == "IsDel" {
		forwardObj.IsDel = arg.Vlaue
	}
	if arg.UpdateType == "ForwardTable" {
		forwardTableList := strings.Split(arg.Vlaue, ";")
		forwardTableData := make([]string, 0)
		for _, v := range forwardTableList {
			if v != "" {
				forwardTableData = append(forwardTableData, v)
			}
		}
		forwardObj.ForwardTable = forwardTableData
	}
	err = new(dao.DaoTCPForward).Set(arg.Name, forwardObj)
	if err != nil {
		rse.Rse = err.Error()
	}
	if rse.Rse == "" {
		rse.Rse = "成功"
	}

R:
	buf, err := protocol.DataEncoder(rse)
	if err != nil {
		logger.Error(err)
	}
	err = ctx.Send(buf)
	if err != nil {
		logger.Error(err)
	}
}
