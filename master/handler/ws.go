package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upGrader = websocket.Upgrader{
		ReadBufferSize:   1024 * 100,
		WriteBufferSize:  65535,
		HandshakeTimeout: 5 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// WsContext websocket context
type WsContext struct {
	Key  string
	Conn *websocket.Conn
}

var notice sync.Map // 通知的ws连接， 服务端都是广播

// BroadcastNotice 广播 通知 数据
func BroadcastNotice(data []byte) {
	notice.Range(func(k, v interface{}) bool {
		client, ok := v.(*WsContext)
		if ok {
			_ = client.Conn.WriteMessage(websocket.TextMessage, data)
		}
		return true
	})
}

func WebSocketNotice(c *gin.Context) {
	// deadLineTimeOut 超时min
	var deadLineTimeOut = 3
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("ws 连接失败 = ", err)
		return
	}
	client := &WsContext{
		Key:  utils.IDMd5(),
		Conn: conn,
	}
	notice.Store(client.Key, client)
	for {
		// 超时错误
		err = client.Conn.SetReadDeadline(time.Now().Add(time.Duration(deadLineTimeOut) * time.Minute))
		if err != nil {
			logger.Error("连接超时")
			notice.Delete(client.Key)
			return
		}
		// 读取数据
		_, data, err := conn.ReadMessage()
		if err != nil {
			logger.Error("读取数据错误")
			notice.Delete(client.Key)
			return
		}
		logger.Info("WebSocket 读取到的数据 ", data)
	}
}

func WebSocketContainerLog(c *gin.Context) {
	slave := c.Query("slave")
	containerId := c.Query("container")
	// 发起 日志采集
	udpC, ok := protocol.AllUdpClient.RetryGet(slave)
	if udpC == nil || !ok {
		c.JSON(200, slave+" 离线  udpC = nil ")
		return
	}
	requst := utils.IDMd5() // 6854823418404110336
	logger.Info("requst = ", requst)
	packate, err := protocol.Packet(protocol.CMD_ContainerLog, requst, []byte(containerId))
	if err != nil {
		logger.Error(err)
	}
	logger.Info("发送数据: ", packate)
	logger.Info(udpC.Conn, udpC.IP)
	protocol.UDPSend(udpC.Conn, packate)
	protocol.Set(requst)
	// 连接ws
	var deadLineTimeOut = 3
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("ws 连接失败 = ", err)
		return
	}
	tx, err := protocol.Get(requst)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("websocket 请求id 过期"))
		return
	}
	go func() {
		for {
			// 超时错误
			err = conn.SetReadDeadline(time.Now().Add(time.Duration(deadLineTimeOut) * time.Minute))
			if err != nil {
				logger.Info("连接超时")
				tx.Stop <- 1
				return
			}
			// 读取数据
			_, data, err := conn.ReadMessage()
			if err != nil {
				tx.Stop <- 1
				return
			}
			logger.Info("WebSocket 读取到的数据 ", data)
		}
	}()
	for {
		select {
		case txErr := <-tx.Err:
			if txErr == nil {
				continue
			}
			_ = conn.WriteMessage(websocket.TextMessage, []byte(txErr.Error()))

		case rse := <-tx.Data:
			if rse == nil {
				continue
			}
			//log.Println(" data = ", rse)
			rseStr := utils.StringValue(rse)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(rseStr))

		case <-tx.Stop:
			//protocol.Close(requst)
			//close(tx.Err)
			//close(tx.Data)
			//close(tx.Stop)
			packate2, err2 := protocol.Packet(protocol.CMD_ContainerLog, requst, []byte("close"))
			if err2 != nil {
				logger.Error(err2)
			}
			logger.Info("发送数据: ", packate2)
			protocol.UDPSend(udpC.Conn, packate2)
			return
		}
	}
}

// WebSocketExecutableLog1 查看可執行文件運行日誌 方案一
func WebSocketExecutableLog1(c *gin.Context) {
	slave := c.Query("slave")
	taskId := c.Query("task_id")
	// 发起 日志采集
	udpC, ok := protocol.AllUdpClient.RetryGet(slave)
	if udpC == nil || !ok {
		c.JSON(200, slave+" 离线  udpC = nil ")
		return
	}
	requst := utils.IDMd5() // 6854823418404110336
	logger.Info("requst = ", requst)
	packate, err := protocol.Packet(protocol.CMD_ExecutablePIDLog, requst, []byte(taskId))
	if err != nil {
		logger.Error(err)
	}
	logger.Info("发送数据: ", packate)
	logger.Info(udpC.Conn, udpC.IP)
	protocol.UDPSend(udpC.Conn, packate)
	protocol.Set(requst)
	// 连接ws
	var deadLineTimeOut = 3
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("ws 连接失败 = ", err)
		return
	}
	tx, err := protocol.Get(requst)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("websocket 请求id 过期"))
		return
	}
	go func() {
		for {
			// 超时错误
			err = conn.SetReadDeadline(time.Now().Add(time.Duration(deadLineTimeOut) * time.Minute))
			if err != nil {
				logger.Info("连接超时")
				tx.Stop <- 1
				return
			}
			// 读取数据
			_, data, err := conn.ReadMessage()
			if err != nil {
				tx.Stop <- 1
				return
			}
			logger.Info("WebSocket 读取到的数据 ", data)
		}
	}()

	for {
		select {
		case txErr := <-tx.Err:
			if txErr == nil {
				continue
			}
			_ = conn.WriteMessage(websocket.TextMessage, []byte(txErr.Error()))

		case rse := <-tx.Data:
			if rse == nil {
				continue
			}
			logger.Info(" data = ", rse)
			rseStr := utils.StringValue(rse)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(rseStr))

		case <-tx.Stop:
			logger.Info("WebSocketExecutableLog tx stop ...... ")
			//protocol.Close(requst)
			//close(tx.Err)
			//close(tx.Data)
			//close(tx.Stop)
			//packate2, err2 := protocol.Packet(protocol.CMD_ContainerLog, requst, []byte("close"))
			//if err2 != nil {
			//	log.Println(err2)
			//}
			//log.Println("发送数据: ", packate2)
			//protocol.UDPSend(udpC.Conn, packate2)
			return
		}
	}
}

// WebSocketExecutableLog 查看可執行文件運行日誌 方案二
func WebSocketExecutableLog(c *gin.Context) {
	slave := c.Query("slave")
	taskId := c.Query("task_id")
	// 发起 日志采集
	udpC, ok := protocol.AllUdpClient.RetryGet(slave)
	if udpC == nil || !ok {
		c.JSON(200, slave+" 离线  udpC = nil ")
		return
	}
	requst := utils.IDMd5() // 6854823418404110336
	protocol.Set(requst)

	// 连接ws
	var deadLineTimeOut = 3
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("ws 连接失败 = ", err)
		return
	}
	tx, err := protocol.Get(requst)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("websocket 请求id 过期"))
		return
	}

	go func() {
		for {
			// 超时错误
			err = conn.SetReadDeadline(time.Now().Add(time.Duration(deadLineTimeOut) * time.Minute))
			if err != nil {
				logger.Error("连接超时")
				tx.Stop <- 1
				return
			}
			// 读取数据
			_, data, err := conn.ReadMessage()
			if err != nil {
				tx.Stop <- 1
				return
			}
			logger.Info("WebSocket 读取到的数据 ", data)
		}
	}()

	go func() {
		for {
			logger.Info("requst = ", requst)
			packate, err := protocol.Packet(protocol.CMD_ExecutablePIDLog, requst, []byte(taskId))
			if err != nil {
				logger.Error(err)
			}
			logger.Info("发送数据: ", packate)
			logger.Info(udpC.Conn, udpC.IP)
			protocol.UDPSend(udpC.Conn, packate)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case txErr := <-tx.Err:
			if txErr == nil {
				continue
			}
			_ = conn.WriteMessage(websocket.TextMessage, []byte(txErr.Error()))

		case rse := <-tx.Data:
			if rse == nil {
				continue
			}
			rseStr := utils.StringValue(rse)
			rseStr = strings.Replace(rseStr, "\\x00", "", -1)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(rseStr))

		case <-tx.Stop:
			logger.Info("WebSocketExecutableLog tx stop ...... ")
			return
		}
	}
}
