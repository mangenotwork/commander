package udp_server

import (
	"fmt"
	"net"
	"strings"
	"time"

	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"
)

func RunUDPServer() {
	go func() {
		var err error
		// 监听
		protocol.UDPListener, err = net.ListenUDP("udp", &net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: conf.MasterConf.UdpServer.Prod,
		})
		if err != nil {
			logger.Error(err)
			panic(err)
			return
		}
		logger.Infof("StartUdpServer Local: <%s> \n", protocol.UDPListener.LocalAddr().String())
		// 读取数据
		data := make([]byte, 1024*10)
		for {
			n, remoteAddr, err := protocol.UDPListener.ReadFromUDP(data)
			if err != nil {
				logger.Error("error during read: %s", err)
			}
			ip := strings.Split(remoteAddr.String(), ":")[0]
			//tx := make(chan interface{})

			client := &protocol.UdpClient{
				Conn: remoteAddr,
				IP:   ip,
			}
			_ = protocol.AllUdpClient.Set(client, 31*time.Second)
			if n > 25 {
				packet := data[:n]
				s := protocol.DecryptPacket(packet, n)
				logger.Infof("[UDP Packet]slave:%s; sign:%s; ctx:%s| %s;",
					client.IP, s.Sign, s.CtxId, s.Command.Chinese())
				go Handle(client, s)
			} else {
				logger.Infof("传入的数据太小或太大, 建议 25~10240个字节")
			}
		}
	}()
}

type UDPHandler map[protocol.CommandCode]func(ctx *HandlerCtx)

// HandlerCtx udp slave 的 Handler 的上下文
type HandlerCtx struct {
	Cmd    protocol.CommandCode
	IP     string
	Stream *protocol.Stream
}

// Send 发送数据包给 master
func (handler *HandlerCtx) Send(data []byte) (string, error) {
	udpC, ok := protocol.AllUdpClient.Get(handler.IP)
	if !ok {
		logger.Infof("192.168.0.9  离线 ok is false")
	}
	logger.Infof("udpC = ", udpC)
	if udpC == nil {
		logger.Infof("192.168.0.9  离线  udpC = nil ")
		return "", fmt.Errorf("udp is null")
	}
	requst := utils.IDMd5() // 6854823418404110336
	logger.Infof("requst = ", requst)
	packate, err := protocol.Packet(protocol.CMD_ReportHostInfo, requst, data)
	if err != nil {
		logger.Error(err)
	}
	logger.Infof("发送数据: ", packate)
	logger.Info(udpC.Conn, udpC.IP)
	protocol.UDPSend(udpC.Conn, packate)
	protocol.Set(requst)
	return requst, nil
}

func Handle(c *protocol.UdpClient, s *protocol.Stream) {
	if f, ok := handle[s.Command]; ok {
		f(&HandlerCtx{
			IP:     c.IP,
			Cmd:    s.Command,
			Stream: s,
		})
	} else {
		logger.Infof("没有找到命令 ", s.Command)
	}
}
