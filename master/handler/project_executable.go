package handler

import (
	"fmt"
	"mime/multipart"
	"path"
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

// ProjectExecutableCreate
// 项目管理 - 二进制物理机部署项目
// 是可执行文件的扩展，是以项目为单位，可以理解为是以目录为单位
// 将可执行文件改为压缩的目录， 新增执行命令， 端口守护等
// 可执行多个副本，均摊部署在不同的salve上
// 可执行项目
// 1. 可执行文件
// 2. 其他文件，如配置，数据等文件
// 3. 执行命令
// 4. 环境变量
// 5. 占用端口
func ProjectExecutableCreate(c *gin.Context) {
	executableFile, err := c.FormFile("executable_file")
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	executableName := c.Request.PostFormValue("executable_name")
	executableNote := c.Request.PostFormValue("executable_note")
	executablePort := c.Request.PostFormValue("executable_port")
	executableCmd := c.Request.PostFormValue("executable_cmd")
	executableEnv := c.Request.PostFormValue("executable_env")
	executableDuplicate := c.Request.PostFormValue("executable_duplicate")
	executableSys := c.Request.PostFormValue("executable_sys")
	if len(executableName) == 0 || len(executableCmd) == 0 {
		APIOutPut(c, 1, 0, "", "参数不全")
		return
	}
	// 保存项目文件
	executablePath, err := saveProjectExecutableFile(c, executableFile)
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	// 解压项目文件
	dst := path.Join(conf.MasterConf.ProjectPath, executableName)
	fileNameWithSuffix := path.Base(executableFile.Filename)
	fileType := path.Ext(fileNameWithSuffix)
	fileNameOnly := strings.TrimSuffix(fileNameWithSuffix, fileType)
	dst = path.Join(dst, fileNameOnly)
	err = utils.DecompressionZipFile(executablePath, dst)
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	projectExecutable := &entity.ProjectExecutable{
		Name:        executableName,
		Note:        executableNote,
		Port:        executablePort,
		Cmd:         executableCmd,
		Env:         executableEnv,
		Duplicate:   executableDuplicate,
		Sys:         executableSys,
		Path:        dst,
		ZipFilePath: executablePath,
		CreateTime:  utils.NowTimeStr(),
	}
	err = new(dao.DaoProjectExecutable).Set(executableName, projectExecutable)
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	APIOutPut(c, 1, 1, "", "创建成功")
	return
}

// saveProjectExecutableFile
func saveProjectExecutableFile(c *gin.Context, file *multipart.FileHeader) (string, error) {
	dst := path.Join(conf.MasterConf.ProjectPath, file.Filename)
	isHave, _ := utils.PathExists(dst)
	if isHave {
		return dst, fmt.Errorf("项目文件已经存在，你新建这个项目以版本号区分!")
	}
	return dst, c.SaveUploadedFile(file, dst)
}

func ProjectExecutableList(c *gin.Context) {
	data, _ := new(dao.DaoProjectExecutable).GetALL()
	APIOutPut(c, 1, 1, data, "创建成功")
	return
}

func ProjectExecutableTaskList(c *gin.Context) {
	projectName := c.Query("project") // ip
	data, _, err := new(dao.DaoProjectExecutable).GetProjectExecutableProcess(projectName)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 1, 0, data, "")
	return
}

func ProjectExecutableRun(c *gin.Context) {
	projectName := c.Query("project")
	project, err := new(dao.DaoProjectExecutable).Get(projectName)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		deployExecutable(project, v)
	}
	APIOutPut(c, 1, 0, "部署已经执行，请关注通知", "ok")
}

// deployExecutable
func deployExecutable(project *entity.ProjectExecutable, slave string) {
	taskId := utils.IDMd5()
	// 记录任务
	new(dao.DaoTask).SetDefaultCreate(slave, taskId, "部署可执行文件项目:"+project.Name)

	// 下发任务
	//fileList := strings.Split(project.ZipFilePath, "/")
	//file := fileList[len(fileList)-1]

	arg := &entity.ProjectExecutableRunArg{
		ProjectName: project.Name,
		Slave:       slave,
		Port:        project.Port,
		Env:         project.Env,
		TaskId:      taskId,
		Cmd:         project.Cmd,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		logger.Error("部署dokcer失败：", err)
		return
	}
	_, err = UDPSend(slave, protocol.CMD_ProjectExecutableRun, buf)
	if err != nil {
		logger.Error("部署dokcer失败：", err)
		return
	}
	//  ======  记录 task
	task := &entity.Task{
		ID:       taskId,
		IP:       slave,
		State:    enum.TaskStateRun.Value(),
		StateStr: enum.TaskStateRun.Str(),
		Create:   utils.NowTimeStr(),
	}
	err = new(dao.DaoTask).Set(taskId, task)
	if err != nil {
		logger.Error("部署dokcer失败：", err)
		return
	}
	logger.Info("docker run 任务启动成功 任务id = ", taskId)
	return
}

func ProjectExecutableDownload(c *gin.Context) {
	projectName := c.Query("project")
	project, err := new(dao.DaoProjectExecutable).Get(projectName)
	if err != nil {
		APIOutPut(c, 1, 1, "", err.Error())
		return
	}
	fileNameList := strings.Split(project.ZipFilePath, "/")
	fileName := fileNameList[len(fileNameList)-1]
	fileContentDisposition := "attachment;filename=\"" + fileName + "\""
	c.Writer.Header().Add("Content-Disposition", fileContentDisposition)
	c.File(project.ZipFilePath)
	return
}
