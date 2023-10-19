package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

func SlaveList(c *gin.Context) {
	data, _ := new(dao.DaoSlave).GetALL()
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
	}
	for _, v := range data {
		_, ok := onlineMap[v.Slave]
		if ok {
			v.SlaveOnline = "在线"
		} else {
			v.SlaveOnline = "离线"
		}
		// 获取当前 cpu 和 内存使用情况
		v.NowCPURate = "0"
		v.NowMEMRate = "0"
		p, err := new(dao.DaoPerformance).GetNowPerformance(v.Slave)
		if err == nil {
			if p.CPU == nil {
				v.NowCPURate = "0"
			} else {
				v.NowCPURate = utils.StringValue(int(p.CPU.UseRate))
			}
			performanceMEM := 0
			if p.MEM != nil {
				performanceMEM = int(float64(p.MEM.MemUsed) / float64(p.MEM.MemTotal) * 100)
			}
			v.NowMEMRate = utils.StringValue(performanceMEM)
		}
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].SlaveOnline < data[j].SlaveOnline
	})
	APIOutPut(c, 0, len(data), data, "")
}

// SlaveProcessList slave 进程列表
func SlaveProcessList(c *gin.Context) {
	slave := c.Query("slave") // ip
	pg := c.Query("pg")       // ip
	UDPSendOutHttp(c, slave, protocol.CMD_SlaveProcessList, []byte(pg))
}

func SlaveENVList(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_SlaveENVList, []byte(""))
}

func SlaveDiskInfo(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_SlaveDiskInfo, []byte(""))
}

func SlavePathInfo(c *gin.Context) {
	slave := c.Query("slave") // ip
	path := c.Query("path")   // ip
	UDPSendOutHttp(c, slave, protocol.CMD_SlavePathInfo, []byte(path))
}

func SlaveProcessKill(c *gin.Context) {
	slave := c.Query("slave") // ip
	pId := c.Query("pid")
	value := c.Query("value")
	if pId == "" {
		APIOutPut(c, 1, 0, "", "pid 不能爲空")
		return
	}
	arg := &entity.ProcessKillArg{
		PID:   pId,
		Value: value,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_ProcessKill, buf)
}

// SlaveProcessInfo 查看进程信息
func SlaveProcessInfo(c *gin.Context) {
	slave := c.Query("slave") // ip
	pId := c.Query("pid")
	UDPSendOutHttp(c, slave, protocol.CMD_SlaveProcessInfo, []byte(pId))
}

// SlavePortInfo slave 端口使用情況
func SlavePortInfo(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_SlavePortInfo, []byte(""))
}

// SlaveHosts 读取 hosts 文件
func SlaveHosts(c *gin.Context) {
	slave := c.Query("slave") // ip
	UDPSendOutHttp(c, slave, protocol.CMD_SlaveHosts, []byte(""))
}

type SlaveHostsUpdateParam struct {
	Slave     string `json:"slave"`
	HostsData string `json:"hosts_data"`
}

// SlaveHostsUpdate 更新 hosts文件
func SlaveHostsUpdate(c *gin.Context) {
	param := &SlaveHostsUpdateParam{}
	err := c.BindJSON(param)
	if err != nil {
		logger.Infof("json err = ", err)
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	taskId := utils.IDMd5()
	arg := entity.SlaveHostsArg{
		TaskId: taskId,
		Data:   param.HostsData,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, param.Slave, protocol.CMD_SlaveHostsUpdate, buf)
}

// SlaveDir slave 目录与文件
func SlaveDir(c *gin.Context) {
	slave := c.Query("slave")
	path := c.Query("path")
	taskId := utils.IDMd5()
	arg := entity.GetSlavePathInfoArg{
		Path:   path,
		TaskId: taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_GetSlavePathInfo, buf)
}

// SlaveCatFile  查看Slave文件内容
func SlaveCatFile(c *gin.Context) {
	slave := c.Query("slave")
	path := c.Query("path")
	taskId := utils.IDMd5()
	arg := entity.CatSlaveFileArg{
		FilePath: path,
		TaskId:   taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_SlaveCatFile, buf)
}

// SlaveFileDownload  下载slave 文件
func SlaveFileDownload(c *gin.Context) {
	slave := c.Query("slave")
	path := c.Query("path")

	slaveObj, err := new(dao.DaoSlave).Get(slave)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	if len(slaveObj.FileServerPort) < 1 {
		APIOutPut(c, 1, 0, "", "没有找到slave file 服务")
		return
	}
	if len(slaveObj.FileServerSecret) < 1 {
		APIOutPut(c, 1, 0, "", "没有找到slave file 服务的secret")
		return
	}

	cli := &http.Client{}
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Print("io.ReadFull(r.Body, body) ", err.Error())
		//return,没有数据也是可以的，不需要直接结束
	}
	fmt.Print("req count :", len(body), "\n")

	reqUrl := "http://" + slave + ":" + slaveObj.FileServerPort + "/file?path=" + path + "&secret=" + slaveObj.FileServerSecret
	logger.Info("下载地址 : ", reqUrl)

	req, err := http.NewRequest(c.Request.Method, reqUrl, strings.NewReader(string(body)))
	if err != nil {
		fmt.Print("http.NewRequest ", err.Error())
		return
	}

	for k, v := range c.Request.Header {
		req.Header.Set(k, v[0])
	}
	res, err := cli.Do(req)
	if err != nil {
		fmt.Print("cli.Do(req) ", err.Error())
		return
	}
	defer res.Body.Close()

	for k, v := range res.Header {
		c.Writer.Header().Set(k, v[0])
	}
	io.Copy(c.Writer, res.Body)
}

// SlaveFileUpload 上传文件到slave
func SlaveFileUpload(c *gin.Context) {
	slave := c.Query("slave")
	slaveObj, err := new(dao.DaoSlave).Get(slave)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	if len(slaveObj.FileServerPort) < 1 {
		APIOutPut(c, 1, 0, "", "没有找到slave file 服务")
		return
	}
	if len(slaveObj.FileServerSecret) < 1 {
		APIOutPut(c, 1, 0, "", "没有找到slave file 服务的secret")
		return
	}

	cli := &http.Client{}
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Print("io.ReadFull(r.Body, body) ", err.Error())
		//return,没有数据也是可以的，不需要直接结束
	}
	fmt.Print("req count :", len(body), "\n")

	reqUrl := "http://" + slave + ":" + slaveObj.FileServerPort + "/set/file?secret=" + slaveObj.FileServerSecret
	logger.Info("下载地址 : ", reqUrl)

	req, err := http.NewRequest(c.Request.Method, reqUrl, strings.NewReader(string(body)))
	if err != nil {
		fmt.Print("http.NewRequest ", err.Error())
		return
	}

	for k, v := range c.Request.Header {
		req.Header.Set(k, v[0])
	}
	res, err := cli.Do(req)
	if err != nil {
		fmt.Print("cli.Do(req) ", err.Error())
		return
	}
	defer res.Body.Close()

	for k, v := range res.Header {
		c.Writer.Header().Set(k, v[0])
	}
	io.Copy(c.Writer, res.Body)
}

// SlaveMkdir 创建目录
func SlaveMkdir(c *gin.Context) {
	slave := c.Query("slave")
	path := c.Query("path")
	taskId := utils.IDMd5()
	arg := entity.SlaveMkdirArg{
		Path:   path,
		TaskId: taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_SlaveMkdir, buf)
}

// SlavePackDir 打包下载整个文件
func SlavePackDir(c *gin.Context) {
	slave := c.Query("slave")
	path := c.Query("path")

	slaveObj, err := new(dao.DaoSlave).Get(slave)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	if len(slaveObj.FileServerPort) < 1 {
		APIOutPut(c, 1, 0, "", "没有找到slave file 服务")
		return
	}
	if len(slaveObj.FileServerSecret) < 1 {
		APIOutPut(c, 1, 0, "", "没有找到slave file 服务的secret")
		return
	}

	cli := &http.Client{}
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Print("io.ReadFull(r.Body, body) ", err.Error())
		//return,没有数据也是可以的，不需要直接结束
	}
	fmt.Print("req count :", len(body), "\n")

	reqUrl := "http://" + slave + ":" + slaveObj.FileServerPort + "/zip/file?path=" + path + "&secret=" + slaveObj.FileServerSecret
	logger.Info("下载地址 : ", reqUrl)

	req, err := http.NewRequest(c.Request.Method, reqUrl, strings.NewReader(string(body)))
	if err != nil {
		fmt.Print("http.NewRequest ", err.Error())
		return
	}

	for k, v := range c.Request.Header {
		req.Header.Set(k, v[0])
	}
	res, err := cli.Do(req)
	if err != nil {
		fmt.Print("cli.Do(req) ", err.Error())
		return
	}
	defer res.Body.Close()

	for k, v := range res.Header {
		c.Writer.Header().Set(k, v[0])
	}
	io.Copy(c.Writer, res.Body)

}

// SlaveDecompress 解压slave 的文件
func SlaveDecompress(c *gin.Context) {
	slave := c.Query("slave")
	path := c.Query("path")
	taskId := utils.IDMd5()
	arg := entity.SlaveDecompressArg{
		Path:   path,
		TaskId: taskId,
	}
	buf, err := protocol.DataEncoder(arg)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	UDPSendOutHttp(c, slave, protocol.CMD_SlaveDecompress, buf)
}
