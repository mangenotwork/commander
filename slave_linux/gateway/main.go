/*
	目前网关功能:
		L4转发
		负载均衡 - 随机发
		TODO:
		黑白名单
		ip限流
*/

package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/slve_linux/dao"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"gitee.com/mangenotework/commander/common/cmd"
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"

	"golang.org/x/sys/unix"
)

var ipsMap map[string][]string
var ipsLock sync.RWMutex

var lc = net.ListenConfig{
	Control: func(network, address string, c syscall.RawConn) error {
		var opErr error
		if err := c.Control(func(fd uintptr) {
			opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		}); err != nil {
			return err
		}
		return opErr
	},
	KeepAlive: 0,
}

// SetMaxOpenFile 设置为最大 socket open file
func SetMaxOpenFile() error {
	var rLimit syscall.Rlimit

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}

	rLimit.Cur = rLimit.Max
	return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
}

func RunGateway(arg *entity.GatewayArg) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	SetMaxOpenFile()

	// 解决端口不够用的问题，connect: cannot assign requested address
	// 调低端口释放后的等待时间，默认为60s，修改为15~30s：
	cmd.LinuxSendCommand("sysctl -w net.ipv4.tcp_fin_timeout=22")
	// 修改tcp/ip协议配置， 通过配置/proc/sys/net/ipv4/tcp_tw_resue, 默认为0，修改为1，释放TIME_WAIT端口给新连接使用：
	cmd.LinuxSendCommand("sysctl -w net.ipv4.tcp_timestamps=1")
	// 修改tcp/ip协议配置，快速回收socket资源，默认为0，修改为1：
	cmd.LinuxSendCommand("sysctl -w net.ipv4.tcp_tw_recycle=1")
	// 允许端口重用：
	cmd.LinuxSendCommand("sysctl -w net.ipv4.tcp_tw_reuse = 1")
	// 全连接队列有空位时，接受到客户端的重试ACK，任然会触发服务端连接成功。
	cmd.LinuxSendCommand("sysctl -w net.ipv4.tcp_abort_on_overflow = 0")

	ipsMap = make(map[string][]string)
	logger.Info("启动网关......")
	rand.Seed(time.Now().UnixNano())

	logger.Info("arg.Ports  = ", arg.Ports)

	for _, v := range arg.Ports {
		portList := strings.Split(v, ":")
		port := portList[0]
		if len(portList) < 2 {
			logger.Info("启动网关失败， 没有映射地址!")
			return
		}
		targetPort := portList[1]
		// 定时拉取注册地址
		go func() {
			for {
				GetIps(arg.ProjectName, targetPort)
				time.Sleep(10 * time.Second)
			}
		}()
		// 启动5个服务
		for i := 0; i < 5; i++ {
			ser := &GatewayServer{
				ServerId:  utils.IDMd5(),
				Port:      port,
				Project:   arg.ProjectName,
				TargePort: targetPort,
				Stop:      make(chan int),
			}
			SetGatewayServer(arg.ProjectName, ser)
			go ser.Run()
			//go server(port, arg.ProjectName, targetPort)
		}
	}
}

var AllGatewayServer map[string][]*GatewayServer = make(map[string][]*GatewayServer)

func SetGatewayServer(project string, ser *GatewayServer) {
	logger.Info("SetGatewayServer ...... ", project, ser)
	if _, ok := AllGatewayServer[project]; !ok {
		AllGatewayServer[project] = make([]*GatewayServer, 0)
	}
	AllGatewayServer[project] = append(AllGatewayServer[project], ser)
}

func CloseGatewayServer(project string) {
	logger.Info("CloseGatewayServer")
	if _, ok := AllGatewayServer[project]; !ok {
		logger.Info("网关不存在")
		return
	}
	logger.Info("AllGatewayServer = ", AllGatewayServer[project])
	for _, v := range AllGatewayServer[project] {
		logger.Info("准备关闭网关 ", v.ServerId)
		go func(v *GatewayServer) {
			v.Stop <- 0
		}(v)
	}
}

type GatewayServer struct {
	ServerId  string
	Port      string
	Project   string
	TargePort string
	Stop      chan int
	Lis       net.Listener
}

func (g *GatewayServer) Run() {
	var err error
	var ip = "0.0.0.0:" + g.Port
	logger.Info("网关地址: ", ip)
	g.Lis, err = lc.Listen(context.Background(), "tcp", ip)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer g.Lis.Close()

	//var sconn net.Conn
	//var dconn net.Conn

	for {
		sconn, err := g.Lis.Accept()
		if err != nil {
			logger.Info("建立连接错误:%v\n", err)
			break
		}
		//logger.Info("连接流向 -> ", conn.RemoteAddr(), conn.LocalAddr())
		go g.handle(context.Background(), sconn, g.Project, g.TargePort)
	}

}

func (g *GatewayServer) handle(ctx context.Context, sconn net.Conn, project, targetPort string) {
	var err error
	ipsLock.RLock()
	ips := ipsMap[fmt.Sprintf("%s%s", project, targetPort)]
	ipsLock.RUnlock()
	logger.Info("获取地址 : ", fmt.Sprintf("%s%s", project, targetPort), ips)
	if len(ips) == 0 {
		// 未获取到地址到持久化获取
		logger.Info("未获取到地址到持久化获取")
		GetPersistenceIps(project, targetPort)
		handle(ctx, sconn, project, targetPort)
		return
	}

	// TODO 负载均衡
	// 随机ip
	ip := ips[rand.Intn(len(ips))]

	// TODO 这里可以使用池化技术进行优化
	dconn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Printf("连接%v失败:%v\n", ip, err)
		return
	}

	//go func(sconn net.Conn, dconn net.Conn) {
	//	//logger.Info("发送数据 ", dconn.RemoteAddr().String(), sconn.RemoteAddr().String())
	//	_, err1 := io.Copy(dconn, sconn)
	//	if err1 != nil {
	//		fmt.Printf("往%v发送数据失败:%v\n", ip, err1)
	//	}
	//}(sconn, dconn)
	//
	//go func(sconn net.Conn, dconn net.Conn) {
	//	//logger.Info("接收数据 ", sconn.RemoteAddr().String(), dconn.RemoteAddr().String())
	//	_, err2 := io.Copy(sconn, dconn)
	//	if err2 != nil {
	//		fmt.Printf("从%v接收数据失败:%v\n", ip, err2)
	//	}
	//}(sconn, dconn)

	go func() {
		//logger.Info("发送数据 ", dconn.RemoteAddr().String(), sconn.RemoteAddr().String())
		_, err1 := io.Copy(dconn, sconn)
		if err1 != nil {
			fmt.Printf("往%v发送数据失败:%v\n", ip, err1)
		}
	}()

	go func() {
		//logger.Info("接收数据 ", sconn.RemoteAddr().String(), dconn.RemoteAddr().String())
		_, err2 := io.Copy(sconn, dconn)
		if err2 != nil {
			fmt.Printf("从%v接收数据失败:%v\n", ip, err2)
		}
	}()

	for {
		select {
		case i := <-g.Stop:
			// 关闭网关， 如果是 ESTABLISHED状态 需要等到生成时间过了才会断掉
			logger.Info("结束服务 handle : ", g.ServerId, i)
			g.Lis.Close()
			logger.Info(sconn, dconn)
			//sconn.Close()
			/*
					默认关闭方式中，即 sec < 0 。操作系统会将缓冲区里未处理完的数据都完成处理，再关闭掉连接。
					当 sec > 0 时，操作系统会以与默认关闭方式运行。但是当超过定义的时间 sec 后，如果还没处理完缓存区的数据，在某些操作系统下，
				缓冲区中未完成的流量可能就会被丢弃。
					而 sec == 0 时，操作系统会直接丢弃掉缓冲区里的流量数据，这就是强制性关闭。
			*/
			err = sconn.(*net.TCPConn).SetLinger(0)
			if err != nil {
				logger.Error("Error when setting linger: %s", err)
			} else {
				logger.Info("sconn SetLinger 0")
			}

			err = dconn.(*net.TCPConn).SetLinger(0)
			if err != nil {
				logger.Error("Error when setting linger: %s", err)
			} else {
				logger.Info("dconn SetLinger 0")
			}

			err = sconn.Close()
			if err != nil {
				logger.Info("sconn Close err = ", err.Error())
			}

			err = dconn.Close()
			if err != nil {
				logger.Info("dconn Close err = ", err.Error())
			}

			return
		}
	}

}

func server(port, project, targetPort string) {
	var ip = "0.0.0.0:" + port
	logger.Info("网关地址: ", ip)
	lis, err := lc.Listen(context.Background(), "tcp", ip)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			logger.Info("建立连接错误:%v\n", err)
			continue
		}
		//logger.Info("连接流向 -> ", conn.RemoteAddr(), conn.LocalAddr())
		go handle(context.Background(), conn, project, targetPort)
	}
}

func handle(ctx context.Context, sconn net.Conn, project, targetPort string) {
	//logger.Info("handle......")
	//defer func() {
	//	//logger.Info("关闭sconn")
	//	sconn.Close()
	//}()

	ipsLock.RLock()
	ips := ipsMap[fmt.Sprintf("%s%s", project, targetPort)]
	ipsLock.RUnlock()
	//logger.Info("获取地址 : ",fmt.Sprintf("%s%s",project,targetPort), ips)
	if len(ips) == 0 {
		// 未获取到地址到持久化获取
		logger.Info("未获取到地址到持久化获取")
		GetPersistenceIps(project, targetPort)
		handle(ctx, sconn, project, targetPort)
		return
	}

	// TODO 负载均衡

	// 随机ip
	ip := ips[rand.Intn(len(ips))]
	//logger.Info("请求ip = ", ip)

	// TODO 这里可以使用池化技术进行优化
	dconn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Printf("连接%v失败:%v\n", ip, err)
		return
	}
	//defer dconn.Close()

	//ExitChan := make(chan bool)

	go func(sconn net.Conn, dconn net.Conn) {
		//logger.Info("发送数据 ", dconn.RemoteAddr().String(), sconn.RemoteAddr().String())
		_, err1 := io.Copy(dconn, sconn)
		if err1 != nil {
			fmt.Printf("往%v发送数据失败:%v\n", ip, err1)
		}
		//ExitChan <- true
		//defer dconn.Close()
	}(sconn, dconn)

	go func(sconn net.Conn, dconn net.Conn) {
		//logger.Info("接收数据 ", sconn.RemoteAddr().String(), dconn.RemoteAddr().String())
		_, err2 := io.Copy(sconn, dconn)
		if err2 != nil {
			fmt.Printf("从%v接收数据失败:%v\n", ip, err2)
		}
		//ExitChan <- true
	}(sconn, dconn)

	//<-ExitChan
	//logger.Info("关闭dconn.....")
	//dconn.Close()
	//sconn.Close()
	//dconn.Close()

	//select {
	//case <-ExitChan:
	//	logger.Info("关闭dconn.....")
	//	dconn.Close()
	//	sconn.Close()
	//	return
	//case <-ctx.Done():
	//	fmt.Println("timeout")
	//	handle(ctx, sconn, project, targetPort)
	//	return
	//

}

// ResponseJson 统一接口输出
type ResponseJson struct {
	Code      int64  `json:"code"`
	Msg       string `json:"msg"`
	Body      Body   `json:"body"`
	TimeStamp int64  `json:"timeStamp"`
}

// Body 具体数据模型
type Body struct {
	Count int         `json:"count"`
	Data  interface{} `json:"data"`
}

func GetIps(project, targetPort string) {
	urlStr := conf.SlaveConf.MasterHttp + "/project/docker/container/ips?project=" + project + "&port=" + targetPort
	//logger.Info("urlStr = ", urlStr)

	resp, err := http.Get(urlStr)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	rse := &ResponseJson{}
	err = json.Unmarshal(body, &rse)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	if rse.Code != 0 {
		logger.Info("重新获取")
		time.Sleep(500 * time.Millisecond)
		GetIps(project, targetPort)
	}

	//logger.Info("rse = ", rse.Body)
	if rse.Body.Data == nil {
		logger.Error(err.Error())
		return
	}

	data := rse.Body.Data
	//logger.Info("GetIps 获取到 ips = ", data.(string))

	ips := strings.Split(data.(string), ";")
	ipsLock.Lock()
	ipsMap[fmt.Sprintf("%s%s", project, targetPort)] = ips
	ipsLock.Unlock()
}

func SetIps(key, ip string) {
	ipsLock.Lock()
	if _, ok := ipsMap[key]; !ok {
		ipsMap[key] = make([]string, 0)
	}
	ipsMap[key] = append(ipsMap[key], ip)
	ipsLock.Unlock()
}

func DelIp(key, ip string) {
	ipsLock.Lock()
	if _, ok := ipsMap[key]; ok {
		for i := 0; i < len(ipsMap[key]); i++ {
			if ipsMap[key][i] == ip {
				ipsMap[key] = append(ipsMap[key][:i], ipsMap[key][i+1:]...)
				i--
			}
		}
	}
	ipsLock.Unlock()
}

// TimingPersistenceIps 定期持久化 ips的数据
func TimingPersistenceIps() {
	go func() {
		for {
			ipsLock.RLock()
			for k, v := range ipsMap {
				new(dao.DaoSlaveGatewayIps).Set([]byte(k), v)
			}
			ipsLock.RUnlock()
			time.Sleep(5 * time.Second)
		}
	}()
}

// GetPersistenceIps 获取持久化ips数据
func GetPersistenceIps(project, targetPort string) {
	key := fmt.Sprintf("%s%s", project, targetPort)
	ips, err := new(dao.DaoSlaveGatewayIps).Get([]byte(key))
	if err != nil {
		GetIps(project, targetPort)
		return
	}
	ipsLock.Lock()
	ipsMap[key] = ips
	ipsLock.Unlock()
}
