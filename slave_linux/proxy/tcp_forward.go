package proxy

import (
	"context"
	"fmt"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/slve_linux/dao"
	"golang.org/x/sys/unix"
	"io"
	"math/rand"
	"net"
	"syscall"
)

var tcpForwardLC = net.ListenConfig{
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

func TCPForward(name, port string) {
	var err error
	var ip = "0.0.0.0:"+port
	logger.Info("TCP转发地址: ", ip)
	Lis, err := tcpForwardLC.Listen(context.Background(),"tcp", ip)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer Lis.Close()

	for {
		sconn, err := Lis.Accept()
		if err != nil {
			logger.Info("建立连接错误:%v\n", err)
			break
		}
		logger.Info("连接流向 -> ", sconn.RemoteAddr(), sconn.LocalAddr())
		go TCPForwardHandle(context.Background(), sconn, name)
	}
}

func TCPForwardHandle(ctx context.Context, sconn net.Conn, name string){
	var err error

	tcpForward, err := new(dao.DaoTCPForward).Get(name)
	if err != nil {
		logger.Error(err)
		return
	}
	if tcpForward == nil {
		logger.Error("tcpForward is null")
		return
	}
	// 是否关闭
	if tcpForward.IsClose == "1" {
		logger.Info("转发已停止...")
		return
	}
	// 是否删除
	if tcpForward.IsDel == "1" {
		logger.Info("转发已删除")
	}
	// 取转发地址表，这个数据应该被本地缓存
	ips := tcpForward.ForwardTable
	// 负载均衡  随机ip
	ip := ips[rand.Intn(len(ips))]
	dconn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Printf("连接%v失败:%v\n", ip, err)
		return
	}

	go func() {
		logger.Info("发送数据 ", dconn.RemoteAddr().String(), sconn.RemoteAddr().String())
		_, err1 := io.Copy(dconn, sconn)
		if err1 != nil {
			fmt.Printf("往%v发送数据失败:%v\n", ip, err1)
		}
	}()

	go func() {
		logger.Info("接收数据 ", sconn.RemoteAddr().String(), dconn.RemoteAddr().String())
		_, err2 := io.Copy(sconn, dconn)
		if err2 != nil {
			fmt.Printf("从%v接收数据失败:%v\n", ip, err2)
		}
	}()
}


