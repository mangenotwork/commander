package handler

import (
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

var PageId = utils.IDMd5()

func PGIndex(c *gin.Context) {
	// slave数量
	_, slaveKey := new(dao.DaoSlave).GetALL()
	slaveOnline := protocol.AllUdpClient.GetAllKey()
	// project数量
	_, executableNum := new(dao.DaoProjectExecutable).GetALL()
	_, dockerNum := new(dao.DaoProjectDocker).GetALL()
	// 警报数量
	_, alarm := new(dao.DaoMonitorRule).GetAlarmALL()
	// 网关数量
	_, gateway := new(dao.DaoGateway).GetALL()
	// 可执行文件数量
	_, executable := new(dao.DaoExecutable).GetALL()
	c.HTML(http.StatusOK, "index.html", gin.H{
		"pgid":       utils.IDMd5(),
		"nav":        "home",
		"slave":      fmt.Sprintf("%d/%d", len(slaveOnline), len(slaveKey)),
		"project":    len(executableNum) + len(dockerNum),
		"alarm":      len(alarm),
		"gateway":    len(gateway),
		"executable": len(executable),
	})
}

func PGLogin(c *gin.Context) {
	token, _ := c.Cookie("authenticate")
	if token != "" {
		j := utils.NewJWT()
		err := j.ParseToken(token)
		if err == nil {
			account := j.GetString("account")
			password := j.GetString("password")
			user, err := new(dao.DaoUser).Get()
			if err == nil && account == user.Account && password == user.Password {
				c.Redirect(http.StatusFound, "/home")
				return
			}
		}
	}
	user, _ := new(dao.DaoUser).Get()
	//logger.Info(err, user)
	if user == nil || user.Account == "" {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"pgid": PageId,
		})
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"pgid": PageId,
	})
	return
}

func Register(c *gin.Context) {
	account := c.Request.PostFormValue("account")
	password := c.Request.PostFormValue("password")
	logger.Info(account)
	logger.Info(password)
	if len(account) == 0 || len(password) == 0 {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"pgid": utils.IDMd5(),
		})
		return
	}
	user, err := new(dao.DaoUser).Get()
	if err == nil && user.Account != "" {
		c.HTML(http.StatusOK, "err.html", gin.H{
			"err": "已存在账号无法创建多个账号!",
		})
		return
	} else {
		_ = new(dao.DaoUser).Set(&entity.User{
			Account:  account,
			Password: utils.MD5String(password),
		})
		c.Redirect(http.StatusFound, "/")
		return
	}
}

var LoginIP sync.Map

// LoginIPStore Store save deviceId and conn
func LoginIPStore(ipStr string, times int) {
	LoginIP.Store(ipStr, times)
}

// LoginIPLoad Load get ws context
func LoginIPLoad(ipStr string) int {
	val, ok := LoginIP.Load(ipStr)
	if ok {
		return val.(int)
	}
	return 0
}

// Delete delete conn
func Delete(ipStr string) {
	LoginIP.Delete(ipStr)
}

// Login
// 保证安全方案
// 1. 定期更新账号密码 TODO
// 2. 记录操作  TODO
// 3. 记录请求  TODO
// 4. 白名单  TODO
// 5. 暴力登录检验，一个ip只有五次试错机会，超过拉黑ip
// master 与 slave 安全方案
// 1. 非对称加密 TODO
func Login(c *gin.Context) {
	account := c.Request.PostFormValue("account")
	password := c.Request.PostFormValue("password")
	logger.Info(account)
	logger.Info(password)
	if len(account) == 0 || len(password) == 0 {
		c.HTML(http.StatusOK, "err.html", gin.H{
			"err": "账号密码为空!",
		})
		return
	}
	// 获取ip， ip 计数
	ip := GetIP(c.Request)
	times := LoginIPLoad(ip)
	if times > 5 {
		c.HTML(http.StatusOK, "err.html", gin.H{
			"err": "已经尝试过5次登录未成功，将封锁客户端ip",
		})
		return
	}
	// 验证
	user, err := new(dao.DaoUser).Get()
	if err == nil && account == user.Account && utils.MD5String(password) == user.Password {
		// TODO 验证白名单
		// 颁发jwt
		j := utils.NewJWT()
		j.AddClaims("account", account)
		j.AddClaims("password", user.Password)
		token, _ := j.Token()
		LoginIPStore(ip, 1)
		c.SetCookie("authenticate", token, 60*60*24*7, "/", "", false, true)
		c.Redirect(http.StatusFound, "/home")
		return
	}
	LoginIPStore(ip, times+1)
	c.HTML(http.StatusOK, "err.html", gin.H{
		"err": "账号或密码错误，尝试5次登录未成功，将封锁客户端ip",
	})
	return
}

func ClearToken(c *gin.Context) {
	c.SetCookie("authenticate", "", 60*60*24*7, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

// GetIP 获取ip
// - X-Real-IP：只包含客户端机器的一个IP，如果为空，某些代理服务器（如Nginx）会填充此header。
// - X-Forwarded-For：一系列的IP地址列表，以,分隔，每个经过的代理服务器都会添加一个IP。
// - RemoteAddr：包含客户端的真实IP地址。 这是Web服务器从其接收连接并将响应发送到的实际物理IP地址。 但是，如果客户端通过代理连接，它将提供代理的IP地址。
//
// RemoteAddr是最可靠的，但是如果客户端位于代理之后或使用负载平衡器或反向代理服务器时，它将永远不会提供正确的IP地址，因此顺序是先是X-REAL-IP，
// 然后是X-FORWARDED-FOR，然后是 RemoteAddr。 请注意，恶意用户可以创建伪造的X-REAL-IP和X-FORWARDED-FOR标头。
func GetIP(r *http.Request) (ip string) {
	for _, ip := range strings.Split(r.Header.Get("X-Forward-For"), ",") {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	if ip = r.Header.Get("X-Real-IP"); net.ParseIP(ip) != nil {
		return ip
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return "0.0.0.0"
}

// OutSlaveList 输出slave列表
type OutSlaveList struct {
	Slave       string `json:"slave"`
	Online      string `json:"online"`
	OnlineValue string `json:"online_value"`
}

func PGDockerManage(c *gin.Context) {
	slave := c.Param("slave")
	msg := ""
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
		if slave == "0" {
			slave = v
		}
	}
	_, ok := onlineMap[slave]
	if !ok && len(online) > 0 {
		msg = " 已经离线。"
		//slave = online[0]
	}
	c.HTML(http.StatusOK, "docker_manage.html", gin.H{
		"pgid":  PageId,
		"Slave": slave,
		"Msg":   msg,
		"nav":   "docker_manage",
	})
}

func PGSlaveSelect(c *gin.Context) {
	data := make([]*OutSlaveList, 0)
	_, keys := new(dao.DaoSlave).GetALL()
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
	}
	for _, v := range keys {
		out := &OutSlaveList{
			Slave: v,
		}
		_, ok := onlineMap[v]
		if ok {
			out.Online = v + "[在线]"
			out.OnlineValue = "在线"
		} else {
			out.Online = v + "[离线]"
			out.OnlineValue = "离线"
		}
		data = append(data, out)
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].OnlineValue < data[j].OnlineValue
	})
	APIOutPut(c, 0, 0, data, "")
}

func PGContainerLog(c *gin.Context) {
	slave := c.Query("slave")
	ContainerId := c.Query("container")
	logger.Infof("slave = ", slave)
	logger.Infof("ContainerId = ", ContainerId)
	c.HTML(http.StatusOK, "container_log.html", gin.H{
		"pgid":         PageId,
		"slave":        slave,
		"container_id": ContainerId,
		"nav":          "docker_manage",
	})
}

func PGExecutableLog(c *gin.Context) {
	slave := c.Query("slave")
	TaskId := c.Query("task_id")
	logger.Infof("slave = ", slave)
	logger.Infof("TaskId = ", TaskId)
	c.HTML(http.StatusOK, "executable_log.html", gin.H{
		"pgid":    PageId,
		"slave":   slave,
		"task_id": TaskId,
		"nav":     "executable",
	})
}

func PGCache(c *gin.Context) {
	c.HTML(http.StatusOK, "cache.html", gin.H{
		"pgid": PageId,
		"nav":  "cache",
	})
}

func PGContainerMonitor(c *gin.Context) {
	slave := c.Query("slave")
	ContainerId := c.Query("container")
	c.HTML(http.StatusOK, "container_monitor.html", gin.H{
		"pgid":         PageId,
		"slave":        slave,
		"container_id": ContainerId,
		"nav":          "docker_manage",
	})
}

func PGSlave(c *gin.Context) {
	slave := c.Param("slave")
	msg := ""
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
		if slave == "0" {
			slave = v
		}
	}
	_, ok := onlineMap[slave]
	if !ok && len(online) > 0 {
		msg = " 已经离线。"
		//slave = online[0]
	}
	// 获取最近的 性能
	// 获取当前 cpu 和 内存使用情况
	cpuRate := "0"
	memRate := "0"
	diskRate := "0"
	p, err := new(dao.DaoPerformance).GetNowPerformance(slave)
	if err == nil && p != nil {
		if p.CPU != nil {
			cpuRate = utils.StringValue(int(p.CPU.UseRate))
		}
		if p.MEM != nil {
			performanceMEM := int(float64(p.MEM.MemUsed) / float64(p.MEM.MemTotal) * 100)
			memRate = utils.StringValue(performanceMEM)
		}
		if p.Disk != nil {
			//allD := 0
			var useD float32 = 0
			for _, d := range p.Disk {
				//logger.Info(d.DistUse.Total, d.DistUse.Rate)
				//allD += utils.Num2Int(d.DistTotalMB)
				useD += d.DistUse.Rate
			}
			performanceDisk := int(useD / float32(len(p.Disk)))
			diskRate = utils.StringValue(performanceDisk)
		}
	}
	c.HTML(http.StatusOK, "slave.html", gin.H{
		"pgid":     PageId,
		"slave":    slave,
		"nav":      "slave",
		"Msg":      msg,
		"cpuRate":  cpuRate,
		"memRate":  memRate,
		"diskRate": diskRate,
	})
}

func PGSlaveProcess(c *gin.Context) {
	slave := c.Param("slave")
	msg := ""
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
	}
	_, ok := onlineMap[slave]
	if !ok && len(online) > 0 {
		msg = " 已经离线。"
		//slave = online[0]
	}
	c.HTML(http.StatusOK, "process.html", gin.H{
		"pgid":  PageId,
		"slave": slave,
		"nav":   "slave",
		"Msg":   msg,
	})
}

func PGSlaveEnv(c *gin.Context) {
	slave := c.Param("slave")
	msg := ""
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
	}
	_, ok := onlineMap[slave]
	if !ok && len(online) > 0 {
		msg = " 已经离线。"
		//slave = online[0]
	}
	c.HTML(http.StatusOK, "environment_variable.html", gin.H{
		"pgid":  PageId,
		"slave": slave,
		"nav":   "slave",
		"Msg":   msg,
	})
}

func PGSlavePort(c *gin.Context) {
	slave := c.Param("slave")
	msg := ""
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
	}
	_, ok := onlineMap[slave]
	if !ok && len(online) > 0 {
		msg = " 已经离线。"
		//slave = online[0]
	}
	c.HTML(http.StatusOK, "port.html", gin.H{
		"pgid":  PageId,
		"slave": slave,
		"nav":   "slave",
		"Msg":   msg,
	})
}

func PGExecutableManage(c *gin.Context) {
	c.HTML(http.StatusOK, "executable.html", gin.H{
		"pgid": PageId,
		"nav":  "executable",
	})
}

func PGProjectManage(c *gin.Context) {
	c.HTML(http.StatusOK, "project.html", gin.H{
		"pgid": PageId,
		"nav":  "project",
	})
}

func PGProjectContainer(c *gin.Context) {
	project := c.Param("project")
	logger.Infof("project = ", project)
	c.HTML(http.StatusOK, "project_container.html", gin.H{
		"pgid":    PageId,
		"project": project,
		"nav":     "project",
	})
}

func PGProjectExecutable(c *gin.Context) {
	project := c.Param("project")
	logger.Infof("project = ", project)
	c.HTML(http.StatusOK, "project_executable.html", gin.H{
		"pgid":    PageId,
		"project": project,
		"nav":     "project",
	})
}

func PGTaskPage(c *gin.Context) {
	c.HTML(http.StatusOK, "task.html", gin.H{
		"pgid": PageId,
		"nav":  "task",
	})
}

func PGGateway(c *gin.Context) {
	c.HTML(http.StatusOK, "gateway.html", gin.H{
		"pgid": PageId,
		"nav":  "project",
	})
}

func PGMonitor(c *gin.Context) {
	c.HTML(http.StatusOK, "monitor.html", gin.H{
		"pgid": PageId,
		"nav":  "monitor",
	})
}

func MonitorSlave(c *gin.Context) {
	slave := c.Param("slave")
	c.HTML(http.StatusOK, "slave_monitor.html", gin.H{
		"pgid":  PageId,
		"slave": slave,
		"nav":   "slave",
	})
}

// PGExecutableDir 查看可执行文件的目录结构
func PGExecutableDir(c *gin.Context) {
	name := c.Query("name")
	executable, _ := new(dao.DaoExecutable).Get([]byte(name))
	c.HTML(http.StatusOK, "executable_dir.html", gin.H{
		"pgid":           PageId,
		"executableName": name,
		"nav":            "executable",
		"cmd":            executable.Cmd,
		"env":            strings.Join(executable.Env, ";"),
	})
}

// PGForward 网络转发管理
func PGForward(c *gin.Context) {
	c.HTML(http.StatusOK, "forward.html", gin.H{
		"pgid": PageId,
		"nav":  "forward",
	})
}

// PGNginx Nginx管理
func PGNginx(c *gin.Context) {
	slave := c.Param("slave")
	msg := ""
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
		if slave == "0" {
			slave = v
		}
	}
	_, ok := onlineMap[slave]
	if !ok && len(online) > 0 {
		msg = " 已经离线。"
		//slave = online[0]
	}

	c.HTML(http.StatusOK, "nginx.html", gin.H{
		"pgid":  PageId,
		"nav":   "nginx_manage",
		"msg":   msg,
		"Slave": slave,
	})
}

// PGEnvDeployed 环境部署
func PGEnvDeployed(c *gin.Context) {
	slave := c.Param("slave")
	msg := ""
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
		if slave == "0" {
			slave = v
		}
	}
	_, ok := onlineMap[slave]
	if !ok && len(online) > 0 {
		msg = " 已经离线。"
		//slave = online[0]
	}
	envCheckRes := &entity.EnvDeployedCheckRse{}
	envCheck := entity.EnvDeployedCheckArg{
		Software: []string{"docker", "nginx"},
	}
	// 获取 是否安装  docker  nginx
	buf, err := protocol.DataEncoder(envCheck)
	if err != nil {
		logger.Error("获取 是否安装 失败 err = ", err.Error())
		return
	}
	rse, err := UDPSend(slave, protocol.CMD_EnvDeployedCheck, buf)
	if err != nil {
		logger.Error(err)
	}
	tx, err := protocol.Get(rse)
	if err != nil {
		c.JSON(200, err)
	}
	defer protocol.Close(rse)
	select {
	case err = <-tx.Err:
		c.JSON(200, err)
	case rse := <-tx.Data:
		logger.Info("rse = ", rse.(*entity.EnvDeployedCheckRse))
		envCheckRes = rse.(*entity.EnvDeployedCheckRse)
	}
	c.HTML(http.StatusOK, "env_deployed.html", gin.H{
		"pgid":     PageId,
		"nav":      "env_deployed",
		"Slave":    slave,
		"Msg":      msg,
		"EnvCheck": envCheckRes,
	})
}

// PGDir 目录&文件  功能: 文件上传 下载 删除 新建目录
func PGDir(c *gin.Context) {
	slave := c.Param("slave")
	msg := ""
	onlineMap := make(map[string]struct{})
	online := protocol.AllUdpClient.GetAllKey()
	for _, v := range online {
		onlineMap[v] = struct{}{}
		if slave == "0" {
			slave = v
		}
	}
	_, ok := onlineMap[slave]
	if !ok && len(online) > 0 {
		msg = " 已经离线。"
		//slave = online[0]
	}
	c.HTML(http.StatusOK, "slave_dir.html", gin.H{
		"pgid":  PageId,
		"slave": slave,
		"nav":   "slave",
		"Msg":   msg,
	})
}

func PGSSH(c *gin.Context) {
	ip := c.Query("ip")
	port := c.Query("port")
	user := c.Query("user")
	password := c.Query("password")
	c.HTML(http.StatusOK, "ssh.html", gin.H{
		"pgid":     PageId,
		"ip":       ip,
		"nav":      "slave",
		"port":     port,
		"user":     user,
		"password": password,
	})
}

// Settings Commander设置
func Settings(c *gin.Context) {
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"pgid": utils.IDMd5(),
		"nav":  "home",
	})
}
