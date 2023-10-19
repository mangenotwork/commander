package udp_server

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/enum"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"
	"gitee.com/mangenotework/commander/master/handler"
)

// Hello 接收slave上报的心跳数据
func Hello(ctx *HandlerCtx) {
	performance := &entity.SlavePerformance{}
	err := protocol.DataDecoder(ctx.Stream.Data, &performance)
	if err != nil {
		log.Println(err)
	}
	err = new(dao.DaoPerformance).SetPerformance(ctx.IP, performance)
	if err != nil {
		logger.Error("保存采集性能失败  err = ", err)
		return
	}
	if performance.CPU == nil || performance.MEM == nil || len(performance.Disk) < 1 {
		return
	}
	// 执行监控标准
	rule, err := new(dao.DaoMonitorRule).Get(ctx.IP)
	if err != nil {
		logger.Error("获取监控标准失败: ", err)
	}

	if rule == nil {
		rule = &entity.MonitorRule{
			MaxCPU:        60,
			MaxMem:        60,
			MaxDisk:       60,
			MaxRx:         1024 * 1024,
			MaxTx:         1024 * 1024,
			MaxConnectNum: 1000,
		}
	}

	// cpu是否超标
	MonitorCPU(ctx.IP, performance, float32(rule.MaxCPU))

	// 内存是否超标
	MonitorMEM(ctx.IP, performance, int64(rule.MaxMem))

	// 磁盘存储是否超标
	MonitorDisk(ctx.IP, performance, rule.MaxDisk)

	// 网络流量是否超标
	MonitorNotwork(ctx.IP, performance, rule.MaxRx, rule.MaxTx)

	// 连接数是否超标
	MonitorConnectNum(ctx.IP, performance.ConnectNum, rule.MaxConnectNum)

}

// MonitorCPU cpu是否超标
func MonitorCPU(slave string, performance *entity.SlavePerformance, ruleCPU float32) {
	if ruleCPU == 0 {
		ruleCPU = float32(entity.MonitorRuleMaxMemDefault)
	}
	if performance.CPU.UseRate > ruleCPU {
		logger.Info("cpu 超标")
		//持久化报警
		id := utils.IDStr()
		_ = new(dao.DaoMonitorRule).SetAlarm(id, &entity.Alarm{
			ID:    id,
			Slave: slave,
			Date:  utils.NowTimeStr(),
			Note:  fmt.Sprintf("cpu 超标, 超过目标:%v, 当前为:%v", ruleCPU, performance.CPU.UseRate),
			Lv:    "警报",
		})
		//发送通知
		OnlineNotice(slave, "cpu 超标！ "+utils.StringValue(performance.CPU.UseRate))
	}
}

// MonitorMEM 内存是否超标
func MonitorMEM(slave string, performance *entity.SlavePerformance, ruleMEM int64) {
	if ruleMEM == 0 {
		ruleMEM = int64(entity.MonitorRuleMaxMemDefault)
	}
	performanceMEM := int64(float64(performance.MEM.MemUsed) / float64(performance.MEM.MemTotal) * 100)
	//logger.Info("内存使用: ", performance.MEM.MemUsed, performance.MEM.MemTotal, performanceMEM)

	if performanceMEM > ruleMEM {
		logger.Info("内存 超标")
		//持久化报警
		id := utils.IDStr()
		_ = new(dao.DaoMonitorRule).SetAlarm(id, &entity.Alarm{
			ID:    id,
			Slave: slave,
			Date:  utils.NowTimeStr(),
			Note:  fmt.Sprintf("mem 超标, 超过目标:%d, 当前为:%d", ruleMEM, performanceMEM),
			Lv:    "警报",
		})
		// 发送通知
		OnlineNotice(slave, "内存 超标！ "+utils.StringValue(performance.MEM.MemUsed))
	}
}

// MonitorDisk 磁盘存储是否超标
func MonitorDisk(slave string, performance *entity.SlavePerformance, ruleDisk int) {
	if ruleDisk == 0 {
		ruleDisk = entity.MonitorRuleMaxDiskDefault
	}
	//diskAllTotal := 0
	for _, v := range performance.Disk {
		if strings.Index(v.DiskName, "dev/s") != -1 {
			//logger.Info(v.DiskName, "|", v.DistUse.Rate, "|", v.DistUse.Total, "|", v.DistUse.Free)
			if v.DistUse.Rate > float32(ruleDisk) {
				logger.Info(fmt.Sprintf("磁盘 %s 超标, 超过目标:%d, 当前为:%f", v.DiskName, ruleDisk, v.DistUse.Rate))
				id := utils.IDStr()
				_ = new(dao.DaoMonitorRule).SetAlarm(id, &entity.Alarm{
					ID:    id,
					Slave: slave,
					Date:  utils.NowTimeStr(),
					Note:  fmt.Sprintf("磁盘 %s 超标, 超过目标:%d, 当前为:%f", v.DiskName, ruleDisk, v.DistUse.Rate),
					Lv:    "警报",
				})
				// 发送通知
				OnlineNotice(slave, fmt.Sprintf("磁盘 %s 超标, 超过目标:%d, 当前为:%f", v.DiskName, ruleDisk, v.DistUse.Rate))
			}
		}
	}
}

// MonitorNotwork 网络流量是否超标
func MonitorNotwork(slave string, performance *entity.SlavePerformance, ruleRx, ruleTx int) {
	if ruleTx == 0 {
		ruleTx = entity.MonitorRuleMaxTxDefault
	}
	if ruleRx == 0 {
		ruleRx = entity.MonitorRuleMaxRxDefault
	}
	for _, v := range performance.NetWork {
		if v.Tx > float32(ruleTx*1024) {
			logger.Info("网络tx 超标")
			id := utils.IDStr()
			_ = new(dao.DaoMonitorRule).SetAlarm(id, &entity.Alarm{
				ID:    id,
				Slave: slave,
				Date:  utils.NowTimeStr(),
				Note:  fmt.Sprintf("%s 网络流量TX 超标, 超过目标:%d, 当前为:%f", v.Name, ruleTx, v.Tx),
				Lv:    "警报",
			})
			OnlineNotice(slave, "网络tx 超标！ "+v.Name+" "+utils.StringValue(v.Tx))
		}
		if v.Rx > float32(ruleRx*1024) {
			logger.Info("网络rx 超标")
			id := utils.IDStr()
			_ = new(dao.DaoMonitorRule).SetAlarm(id, &entity.Alarm{
				ID:    id,
				Slave: slave,
				Date:  utils.NowTimeStr(),
				Note:  fmt.Sprintf("%s 网络流量RX 超标, 超过目标:%d, 当前为:%f", v.Name, ruleRx, v.Rx),
				Lv:    "警报",
			})
			OnlineNotice(slave, "网络rx 超标！ "+v.Name+" "+utils.StringValue(v.Rx))
		}
	}
}

// MonitorConnectNum 连接数是否超标
func MonitorConnectNum(slave string, connectNum, ruleConnectNum int) {
	if ruleConnectNum == 0 {
		ruleConnectNum = entity.MonitorRuleMaxConnectNumDefault
	}
	if connectNum > ruleConnectNum {
		logger.Info("连接数 超标")
		id := utils.IDStr()
		_ = new(dao.DaoMonitorRule).SetAlarm(id, &entity.Alarm{
			ID:    id,
			Slave: slave,
			Date:  utils.NowTimeStr(),
			Note:  fmt.Sprintf("主机连接数 超标, 超过目标:%d, 当前为:%d", ruleConnectNum, connectNum),
			Lv:    "警报",
		})
		OnlineNotice(slave, "主机连接数 超标！ "+utils.StringValue(connectNum))
	}
}

// OnlineNotice  发送 在线通知
func OnlineNotice(slave, notice string) {
	handler.BroadcastNotice([]byte("[警报] " + slave + " | " + utils.NowTimeStr() + ":" + notice))
}

func ReportHostInfo(ctx *HandlerCtx) {
	hostInfo := &entity.HostInfo{}
	err := protocol.GobDecoder(hostInfo, ctx.Stream.Data)
	if err != nil {
		log.Println(err)
		return
	}
	hostInfo.Slave = ctx.IP
	logger.Info(" docker 相关信息 :  ", hostInfo.HasDocker, hostInfo.DockerVersion)
	// 数据持久化
	err = new(dao.DaoSlave).Set(ctx.IP, hostInfo)
	if err != nil {
		log.Println("数据持久化 失败 = ", err)
	}
	// 如果slave没有安装docker 环境则进行报警处理
	if hostInfo.HasDocker == "false" {
		id := utils.IDStr()
		_ = new(dao.DaoMonitorRule).SetAlarm(id, &entity.Alarm{
			ID:    id,
			Slave: ctx.IP,
			Date:  utils.NowTimeStr(),
			Note:  fmt.Sprintf("【提醒】%s未安装docker", ctx.IP),
			Lv:    "警报",
		})
		//发送通知
		OnlineNotice(ctx.IP, fmt.Sprintf("【提醒】%s未安装docker", ctx.IP))
	}
	// 如果是slave自行上报就不用输出到http
	if ctx.Stream.CtxId == "0000000000000000" {
		return
	}
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- hostInfo
	}
	return
}

func HaveDocker(ctx *HandlerCtx) {
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err == nil {
		tx.Data <- string(ctx.Stream.Data)
	}
}

func SlaveProcessList(ctx *HandlerCtx) {
	data := make([]*entity.ProcessBaseInfo, 0)
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	log.Println("SlaveProcessList = ", data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

func SlaveENVList(ctx *HandlerCtx) {
	data := make([]string, 0)
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	log.Println("SlaveProcessList = ", data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

func SlaveDiskInfo(ctx *HandlerCtx) {
	data := make([]*entity.DiskInfo, 0)
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	log.Println("SlaveProcessList = ", data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

func SlavePathInfo(ctx *HandlerCtx) {
	data := make([]*entity.FileInfo, 0)
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	log.Println("SlaveProcessList = ", data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

func ExecutableDeploy(ctx *HandlerCtx) {
	data := &entity.ExecutableDeployTask{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	log.Println("ExecutableDeploy = ", data.TaskId, data)
	if err != nil {
		log.Println("ExecutableDeploy err = ", err)
		// 记录任务
		new(dao.DaoTask).SetBreakOffRse(ctx.IP, data.TaskId, " 運行可執行文件 失敗，解析數據失敗;")
		// 下发通知
		handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " 運行可執行文件 失敗，解析數據失敗; " +
			utils.StringValue(data)))
		return
	}
	// 可執行文件執行數據 持久化
	err = new(dao.DaoSlaveExecutableTask).Delete(data.TaskId)
	if err != nil {
		log.Println("删除旧数据失败 : ", data.TaskId)
	}
	err = new(dao.DaoSlaveExecutableTask).Set(data.TaskId, data)
	if err != nil {
		log.Println("持久化數據失敗!")
		// 下发通知
		handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: 持久化數據失敗; " +
			utils.StringValue(data)))
		return
	}
	// 记录任务
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId, "運行可執行文件  成功 pid = "+utils.StringValue(data.PID))
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: 運行可執行文件  成功; " +
		utils.StringValue(data)))
	return
}

func ProjectExecutableRun(ctx *HandlerCtx) {
	data := &entity.ProjectExecutableRunRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	log.Println("ExecutableDeploy = ", data.TaskId, data)
	if err != nil {
		// 记录任务
		new(dao.DaoTask).SetBreakOffRse(ctx.IP, data.TaskId, "運行可執行文件 失敗，解析數據失敗; ")
		// 下发通知
		log.Println("ExecutableDeploy err = ", err)
		handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " 運行可執行文件 失敗，解析數據失敗; " +
			utils.StringValue(data)))
		return
	}
	log.Println("ProjectExecutableRun... = ", data.ProjectName, data.TaskId)
	// TODO 可執行文件项目執行數據 持久化
	err = new(dao.DaoProjectExecutable).SetProjectExecutableProcess(data.ProjectName, data.TaskId,
		&entity.ExecutableProcess{
			Slave:   ctx.IP,
			Project: data.ProjectName,
			PID:     data.Pid,
			TaskID:  data.TaskId,
			Cmd:     data.Cmd,
		})
	if err != nil {
		log.Println("保存项目的部署进程数据失败 err = ", err.Error())
	}
	// 记录任务
	new(dao.DaoTask).SetCompleteRse(ctx.IP, data.TaskId,
		"運行可執行文件  成功 pid = "+"運行可執行文件  成功; pid= "+utils.StringValue(data.Pid))
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + data.TaskId + "]: 運行可執行文件  成功; " +
		utils.StringValue(data)))
	return
}

func GatewayRun(ctx *HandlerCtx) {
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- string(ctx.Stream.Data)
	}
	return
}

func SlavePortInfo(ctx *HandlerCtx) {
	data := make([]*entity.PortInfo, 0)
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

// ProcessKill 關閉進程
func ProcessKill(ctx *HandlerCtx) {
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		str := string(ctx.Stream.Data)
		if str == "" {
			str = "關閉進程成功"
		}
		tx.Data <- str
	}
}

// ExecutableRunState 查看進程狀態
func ExecutableRunState(ctx *HandlerCtx) {
	data := &entity.ExecutableRunStateRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &data)
	rseStr := utils.StringValue(data.Rse)
	rseStr = strings.Replace(rseStr, "\"", "", -1)
	task, taskErr := new(dao.DaoSlaveExecutableTask).Get(data.TaskId)
	if taskErr != nil {
		logger.Infof("task get err : ", taskErr)
		return
	}
	task.State = rseStr
	if rseStr == enum.ExecutableStateDiscontinued {
		task.PID = 0
	}
	taskErr = new(dao.DaoSlaveExecutableTask).Set(data.TaskId, task)
	if taskErr != nil {
		logger.Infof("task set err : ", taskErr)
		return
	}
	tx, err := protocol.Get(data.TaskId)
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- data
	}
	return
}

// GetSlavePathInfo 获取指定路径的目录与文件
func GetSlavePathInfo(ctx *HandlerCtx) {
	rse := &entity.GetSlavePathInfoRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil {
		logger.Error("获取指定路径的目录与文件 err = ", err)
	}
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		logger.Info("rse.FileStructure = ", rse.FileStructure)
		tx.Data <- rse.FileStructure
	}
	return
}

// SlaveCatFile 查看文件内容
func SlaveCatFile(ctx *HandlerCtx) {
	rse := &entity.CatSlaveFileRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil {
		logger.Error("获取指定路径的目录与文件 err = ", err)
	}
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- rse.Data
	}
	return
}

// SlaveMkdir slave创建目录
func SlaveMkdir(ctx *HandlerCtx) {
	rse := &entity.SlaveMkdirRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil {
		logger.Error("slave创建目录 err = ", err)
	}
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- rse.Rse
	}
	return
}

// SlaveDecompress 解压slave文件
func SlaveDecompress(ctx *HandlerCtx) {
	rse := &entity.SlaveDecompressRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil {
		logger.Error("解压slave文件 err = ", err)
	}
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- rse.Rse
	}
	return
}

// NginxInfo 获取nginx信息
func NginxInfo(ctx *HandlerCtx) {
	rse := &entity.NginxInfoRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	tx, err := protocol.Get(ctx.Stream.CtxId)
	if err != nil {
		logger.Error("获取nginx信息 err = ", err)
	}
	if err != nil && tx != nil {
		tx.Err <- err
	} else if tx != nil {
		tx.Data <- rse.Rse
	}
	return
}

// NginxStart 启动nginx
func NginxStart(ctx *HandlerCtx) {
	rse := &entity.NginxStartRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("获取 启动nginx err = ", err)
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId +
		"]: 启动nginx " + rse.Rse))
}

// NginxReload 重启nginx
func NginxReload(ctx *HandlerCtx) {
	rse := &entity.NginxReloadRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("获取 重启nginx err = ", err)
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId +
		"]: 重启nginx " + rse.Rse))
}

// NginxQuit 停止nginx
func NginxQuit(ctx *HandlerCtx) {
	rse := &entity.NginxQuitRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("获取 停止nginx err = ", err)
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId +
		"]: 停止nginx " + rse.Rse))
}

// NginxStop 强制停止nginx
func NginxStop(ctx *HandlerCtx) {
	rse := &entity.NginxStopRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("获取 强制停止nginx err = ", err)
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId +
		"]: 强制停止nginx " + rse.Rse))
}

// NginxCheckConf 检查nginx配置
func NginxCheckConf(ctx *HandlerCtx) {
	rse := &entity.NginxCheckConfRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("获取 强制停止nginx err = ", err)
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId +
		"]: 检查nginx配置 : " + rse.Rse))
}

// NginxConfUpdate 修改nginx配置文件
func NginxConfUpdate(ctx *HandlerCtx) {
	rse := &entity.NginxConfUpdateRse{}
	err := protocol.DataDecoder(ctx.Stream.Data, &rse)
	if err != nil {
		logger.Error("获取 强制停止nginx err = ", err)
	}
	// 下发通知
	handler.BroadcastNotice([]byte(utils.Int642Str(time.Now().Unix()) + " [task: " + rse.TaskId +
		"]: 修改nginx配置文件 : " + rse.Rse))
}
