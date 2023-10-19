package handler

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/enum"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

// ExecutableUpload 上传可执行文件
func ExecutableUpload(c *gin.Context) {
	name := c.Request.PostFormValue("name")
	cmd := c.Request.PostFormValue("cmd")
	env := c.Request.PostFormValue("env")
	osType := c.Request.PostFormValue("os_type")
	logger.Infof("name = ", name)
	logger.Info("cmd = ", cmd)
	logger.Info("env = ", env)
	logger.Infof("osType = ", osType)
	if name == "" {
		APIOutPut(c, 1, 1, "", "名称为空")
		return
	}
	// 是否已经存在
	_, err := new(dao.DaoExecutable).Get([]byte(name))
	if err == nil {
		logger.Infof(" 已经存在 ！！ ")
		APIOutPut(c, 1, 1, "", "名称已经存在，可以使用 NAMEv0.0.1 这样来区分")
		return
	}
	if osType != "win" && osType != "unix" {
		APIOutPut(c, 1, 1, "", "未知系统类型参数")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	// 判断是否是 zip tar 文件
	// 获取文件名称带后缀
	fileNameWithSuffix := path.Base(file.Filename)
	// 获取文件的后缀(文件类型)
	fileType := path.Ext(fileNameWithSuffix)
	if !(fileType == ".zip" || fileType == ".tar") {
		APIOutPut(c, 1, 1, "", "不支持的文件类型")
		return
	}
	// file.Filename是否存在
	dst := path.Join(conf.MasterConf.ExeStoreHousePath, file.Filename)
	isHave, _ := utils.PathExists(dst)
	if isHave {
		//获取文件名称(不带后缀)
		fileNameOnly := strings.TrimSuffix(fileNameWithSuffix, fileType)
		dst = path.Join(conf.MasterConf.ExeStoreHousePath, fileNameOnly+"_"+utils.NowUnixStr()+fileType)
	}
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	newFileNameList := strings.Split(dst, "/")
	newFileName := newFileNameList[len(newFileNameList)-1]
	logger.Info("newFileName = ", newFileName)
	outPath := strings.Split(dst, ".")[0] + "/"
	// 解压文件
	logger.Info("解压文件 : ", dst, " -> ", outPath)
	if fileType == ".zip" {
		err = utils.DeCompressZIP(dst, outPath)
	} else if fileType == ".tar" {
		err = utils.DeCompressTAR(dst, outPath)
	}
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	//// 删除原始文件
	//err = os.Remove(dst)
	//if err != nil {
	//	APIOutPut(c, 1, 1, "", err.Error())
	//	return
	//}
	exeFile := &entity.ExecutableFile{
		Name:         name,
		FileName:     newFileName,
		SaveFilePath: dst,
		Path:         outPath,
		DirPath:      outPath,
		Size:         utils.StringValue(file.Size),
		OSType:       osType,
		UploadTime:   utils.NowTimeStr(),
		FileID:       utils.MD5String(name),
		Cmd:          cmd,
		Env:          strings.Split(env, ";"),
	}
	exeFile.Md5, err = utils.Md5BigFile(dst)
	err = new(dao.DaoExecutable).Set(name, exeFile)
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	APIOutPut(c, 0, 1, "上传成功", "")
	return
}

// ExecutableDownload 下载可执行文件
func ExecutableDownload(c *gin.Context) {
	name := c.Query("name")
	f, err := new(dao.DaoExecutable).Get([]byte(name))
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	fileContentDisposition := "attachment;filename=\"" + f.FileName + "\""
	c.Writer.Header().Add("Content-Disposition", fileContentDisposition)
	c.File(f.SaveFilePath)
	return
}

// ExecutableListOut 可执行文件列表输出
type ExecutableListOut struct {
	List []*entity.ExecutableFile
	Keys []string
}

// ExecutableList 可执行文件列表
func ExecutableList(c *gin.Context) {
	data := &ExecutableListOut{}
	data.List, data.Keys = new(dao.DaoExecutable).GetALL()
	APIOutPut(c, 0, 1, data, "")
	return
}

// ExecutableDelete 删除可执行文件
func ExecutableDelete(c *gin.Context) {
	name := c.Query("name")
	f, _ := new(dao.DaoExecutable).Get([]byte(name))
	err := new(dao.DaoExecutable).Delete(name)
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	if f != nil {
		err = os.RemoveAll(f.DirPath)
		err = os.RemoveAll(f.SaveFilePath)
		if err != nil {
			APIOutPut(c, 1, 1, "", err.Error())
			return
		}
	}
	APIOutPut(c, 1, 1, "删除成功", "")
	return
}

// ExecutableDeploy 部署可执行文件
func ExecutableDeploy(c *gin.Context) {
	name := c.Request.PostFormValue("name")   // 执行文件名称
	cmd := c.Request.PostFormValue("cmd")     // 运行的命令
	env := c.Request.PostFormValue("env")     // 运行的环境变量
	slave := c.Request.PostFormValue("slave") // ip  指定主机 | 如果空就随机主机
	arg := c.Request.PostFormValue("arg")     // 空格隔开
	note := c.Request.PostFormValue("note")   // 備註
	taskId := utils.IDMd5()
	deploy := entity.ExecutableDeployArg{
		Slave:        slave,
		DownloadFile: name,
		Arg:          arg,
		TaskId:       taskId,
		Note:         note,
		Cmd:          cmd,
		Env:          env,
	}
	buf, err := protocol.DataEncoder(deploy)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	// 记录到任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "部署可执行文件: "+name)
	// 下发任务
	_, err = UDPSend(slave, protocol.CMD_ExecutableDeploy, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "提交執行，請等待通知！")
}

// ExecutableRunList 已經執行的可執行文件
func ExecutableRunList(c *gin.Context) {
	data, _ := new(dao.DaoSlaveExecutableTask).GetALL()
	APIOutPut(c, 0, 0, data, "")
}

// ExecutableRunState 查看可執行文件任務的狀態 查看是否在運行
func ExecutableRunState(c *gin.Context) {
	slave := c.Query("slave")    // ip
	pid := c.Query("pid")        //
	taskId := c.Query("task_id") // 可執行文件任務id
	if pid == "0" {
		APIOutPut(c, 1, 0, "", "PID为空，请刷新列表")
		return
	}
	arg := &entity.ExecutableRunStateArg{
		PId:    pid,
		TaskId: taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_ExecutableRunState, buf)
}

// ExecutableRunLog test case 查看已經執行的可執行文件的日誌
func ExecutableRunLog(c *gin.Context) {
	slave := c.Query("slave")    // ip
	taskId := c.Query("task_id") // 可執行文件任務id
	reqId, err := UDPSend(slave, protocol.CMD_ExecutablePIDLog, []byte(taskId))
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	logger.Infof("ExecutableRunState taskId = ", reqId)
	APIOutPut(c, 0, 1, "", "")
}

// ExecutableTaskDelete 刪除已經執行的可执行文件任務  如果正在執行無法刪除
func ExecutableTaskDelete(c *gin.Context) {
	taskId := c.Query("task_id") // 可執行文件任務id
	// 查看是否正在执行
	task, taskErr := new(dao.DaoSlaveExecutableTask).Get(taskId)
	if taskErr != nil {
		logger.Infof("task get err : ", taskErr)
		APIOutPut(c, 1, 0, "", taskErr.Error())
		return
	}
	if task.State == enum.ExecutableStateExecuting {
		APIOutPut(c, 1, 0, "", "无法删除，因为该任务正在执行中！如果需要删除请kill掉进程")
		return
	}
	err := new(dao.DaoSlaveExecutableTask).Delete(taskId)
	if err != nil {
		logger.Infof("task delete err : ", err)
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 1, "", "删除成功!")
}

// ExecutableTaskRun 启动可执行文件任务 如果已经启动无法再次启动
func ExecutableTaskRun(c *gin.Context) {
	slave := c.Query("slave")    // ip
	taskId := c.Query("task_id") // 可執行文件任務id
	// 查看是否正在执行
	task, taskErr := new(dao.DaoSlaveExecutableTask).Get(taskId)
	if taskErr != nil {
		logger.Infof("task get err : ", taskErr)
		APIOutPut(c, 1, 0, "", taskErr.Error())
		return
	}
	if task.State == enum.ExecutableStateExecuting {
		APIOutPut(c, 1, 0, "", "无法启动，因为该任务正在执行中！")
		return
	}
	deploy := entity.ExecutableDeployArg{
		Slave:        task.Slave,
		DownloadFile: task.ExecutableName,
		Arg:          task.Arg,
		TaskId:       taskId,
		Note:         task.Note,
		Cmd:          task.Command,
		Env:          task.Env,
	}
	buf, bufErr := protocol.DataEncoder(deploy)
	if bufErr != nil {
		APIOutPut(c, 1, 0, "", bufErr.Error())
		return
	}
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "部署可执行文件:"+task.ExecutableName)
	// 下发任务
	_, err := UDPSend(slave, protocol.CMD_ExecutableDeploy, buf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "提交執行，請等待通知！")
}

// ExecutableTaskRestart 重啓已經執行的可执行文件進程
func ExecutableTaskRestart(c *gin.Context) {
	ExecutableTaskDelete(c)
	ExecutableTaskRun(c)
}

// ExecutableKill 停止已經執行的可執行文件
func ExecutableKill(c *gin.Context) {
	slave := c.Query("slave") // ip
	pId := c.Query("pid")
	taskId := c.Query("task_id") // 可執行文件任務id
	value := c.Query("value")
	if pId == "" {
		APIOutPut(c, 1, 0, "", "pid 不能爲空")
		return
	}
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "kill掉可执行文件pid="+pId)
	// 下发任务
	arg := &entity.ExecutableKillArg{
		PID:    pId,
		Value:  value,
		TaskId: taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_ExecutableKill, buf)
}

// ExecutableTaskPIDInfo 查看进程详情  没有执行则无法查看
func ExecutableTaskPIDInfo(c *gin.Context) {
	SlaveProcessInfo(c)
}

// MonitorRuleList  监控标准列表
func MonitorRuleList(c *gin.Context) {
	data, _ := new(dao.DaoMonitorRule).GetALL()
	APIOutPut(c, 0, 0, data, "")
}

// MonitorRuleCreate 新增监控标准
func MonitorRuleCreate(c *gin.Context) {
	slave := c.Request.PostFormValue("slave")
	maxCPU := c.Request.PostFormValue("max_cpu")
	maxMem := c.Request.PostFormValue("max_mem")
	maxDisk := c.Request.PostFormValue("max_disk")
	maxTx := c.Request.PostFormValue("max_tx")
	maxRx := c.Request.PostFormValue("max_rx")
	maxConnectNum := c.Request.PostFormValue("max_connect_num")
	err := new(dao.DaoMonitorRule).Set(slave, &entity.MonitorRule{
		Slave:         slave,
		MaxCPU:        utils.Num2Int(maxCPU),
		MaxMem:        utils.Num2Int(maxMem),
		MaxDisk:       utils.Num2Int(maxDisk),
		MaxTx:         utils.Num2Int(maxTx),
		MaxRx:         utils.Num2Int(maxRx),
		MaxConnectNum: utils.Num2Int(maxConnectNum),
	})
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "成功")
}

func MonitorAlarmList(c *gin.Context) {
	pg := c.Query("pg")
	data := new(dao.DaoMonitorRule).GetAlarmALLPage(utils.Any2Int(pg))
	APIOutPut(c, 0, 0, data, "成功")
}

func MonitorAlarmDel(c *gin.Context) {
	id := c.Query("id")
	err := new(dao.DaoMonitorRule).DeleteAlarm(id)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "成功")
}

// OutMonitorData  输出监控采集的数据
type OutMonitorData struct {
	Date     string
	CPU      float32
	MEM      int
	Tx       float32
	Rx       float32
	MEMUsage int64
	MEMLimit int64
	Connect  int
}

func MonitorData(c *gin.Context) {
	slave := c.Query("slave")
	data := make([]*OutMonitorData, 0)
	// 获取最近一天的数据
	performance, _ := new(dao.DaoPerformance).GetPerformanceMinute(slave)
	for _, p := range performance {
		if p.CPU == nil {
			continue
		}
		d := &OutMonitorData{}
		d.Date = p.TimeStamp
		d.CPU = p.CPU.UseRate
		performanceMEM := int(float64(p.MEM.MemUsed) / float64(p.MEM.MemTotal) * 100)
		d.MEM = performanceMEM
		d.MEMUsage = p.MEM.MemUsed
		d.MEMLimit = p.MEM.MemTotal
		d.Tx = p.NetWork[0].Tx
		d.Rx = p.NetWork[0].Rx
		d.Connect = p.ConnectNum
		data = append(data, d)
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date < data[j].Date
	})
	APIOutPut(c, 0, 0, data, "成功")
}

// ExecutableDir 查看可执行文件的目录
func ExecutableDir(c *gin.Context) {
	name := c.Query("name")
	executable, err := new(dao.DaoExecutable).Get([]byte(name))
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	// 获取目录结构
	files, err := ioutil.ReadDir(executable.DirPath)
	if err != nil {
		log.Println(err)
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	fileStructure := make([]*ExecutableFileStructure, 0)
	for _, v := range files {
		filename := v.Name()
		isDir := v.IsDir()
		fileNameWithSuffix := path.Base(filename)
		//获取文件的后缀(文件类型)
		isConf := false
		fileType := path.Ext(fileNameWithSuffix)
		if fileType == ".conf" || fileType == ".yaml" || fileType == ".ini" {
			isConf = true
		}
		fileStructure = append(fileStructure, &ExecutableFileStructure{
			FileName: filename,
			IsDir:    isDir,
			IsConf:   isConf,
		})
	}
	APIOutPut(c, 0, 0, fileStructure, "")
	return
}

// ExecutableFileStructure 可执行文件的目录文件结构输出实体
type ExecutableFileStructure struct {
	FileName string
	IsDir    bool
	IsConf   bool
}

// ExecutableConfFile 打开配置文件
func ExecutableConfFile(c *gin.Context) {
	name := c.Query("name")
	file := c.Query("file")
	executable, err := new(dao.DaoExecutable).Get([]byte(name))
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	logger.Info(executable.DirPath + file)
	content, err := os.ReadFile(executable.DirPath + file)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, string(content), "")
}

// ExecutableConfUpdateParam 更新可执行文件配置文件实体
type ExecutableConfUpdateParam struct {
	Name     string `json:"name"`
	ConfName string `json:"conf_name"`
	NewConf  string `json:"new_conf"`
}

// ExecutableConfUpdate  修改配置文件
func ExecutableConfUpdate(c *gin.Context) {
	param := &ExecutableConfUpdateParam{}
	err := c.BindJSON(param)
	if err != nil {
		logger.Infof("json err = ", err)
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	executable, err := new(dao.DaoExecutable).Get([]byte(param.Name))
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	confFilePath := executable.DirPath + param.ConfName
	logger.Info("confFilePath = ", confFilePath)
	// O_WRONLY: 只写, O_TRUNC: 清空文件
	file, err := os.OpenFile(confFilePath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("文件打开错误", err)
		return
	}
	defer file.Close() // 关闭文件
	// 带缓冲区的*Writer
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(param.NewConf)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	// 将缓冲区中的内容写入到文件里
	err = writer.Flush()
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	allFile, err := utils.GetAllFile(executable.DirPath)
	if err != nil {
		APIOutPut(c, 1, 0, "", "打包压缩失败： "+err.Error())
		return
	}
	// 打包并替换 zip文件
	err = utils.Compress(allFile, executable.SaveFilePath)
	if err != nil {
		APIOutPut(c, 1, 0, "", "打包压缩失败： "+err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "成功")
}

// TODO ExecutableTaskLog 查看可执行文件运行输出的终端打印日志
func ExecutableTaskLog(c *gin.Context) {
	slave := c.Query("slave")
	taskId := c.Query("task_id")
	logData, err := UDPSend(slave, protocol.CMD_ExecutablePIDLog, []byte(taskId))
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	logger.Info("logData = ", logData)
	tx, err := protocol.Get(logData)
	if err != nil {
		c.JSON(200, err)
	}
	defer protocol.Close(logData)
	select {
	case err = <-tx.Err:
		APIOutPut(c, 0, 0, "", err.Error())
	case rse := <-tx.Data:
		//logger.Infof("protocol.Get(requst) data = ", rse.(string))
		APIOutPut(c, 0, 0, rse.(string), "")
	}
}
