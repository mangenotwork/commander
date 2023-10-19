package protocol

import (
	"net"
	"sync"
	"time"

	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/logger"
)

var (
	AllUdpClient = NewUdpClientCache() // 所有slave 的udp客户端
	UDPListener  *net.UDPConn          // 监听的udp
)

func UDPSend(addr *net.UDPAddr, b []byte) {
	if UDPListener != nil {
		_, _ = UDPListener.WriteToUDP(b, addr)
	}
}

// UdpClient slave udp client
type UdpClient struct {
	Conn       *net.UDPAddr
	IP         string
	Expiration int64 // 过期时间
}

func NewUdpClient() *UdpClient {
	return new(UdpClient)
}

func (udp *UdpClient) AddConn(conn *net.UDPAddr) *UdpClient {
	udp.Conn = conn
	return udp
}

func (udp *UdpClient) SetIP(ip string) *UdpClient {
	udp.IP = ip
	return udp
}

// Expired 判断数据项是否已经过期
func (udp *UdpClient) Expired() bool {
	if udp.Expiration == 0 {
		return false
	}
	return time.Now().Unix() > udp.Expiration
}

// UdpClientCache 内存缓存保存的udp 客户端
type UdpClientCache struct {
	defaultExpiration time.Duration
	items             map[string]*UdpClient // 缓存数据项存储在 map 中
	mu                sync.RWMutex          // 读写锁
	gcInterval        time.Duration         // 过期数据项清理周期
	stopGc            chan bool
}

// NewUdpClientCache 新的udp客户端缓存
func NewUdpClientCache() *UdpClientCache {
	c := &UdpClientCache{
		defaultExpiration: 30 * time.Second,
		gcInterval:        31 * time.Second,
		items:             map[string]*UdpClient{},
		stopGc:            make(chan bool),
	}
	// 开始启动过期清理 goroutine
	go c.gcLoop()
	return c
}

// gcLoop 过期缓存数据项清理
func (c *UdpClientCache) gcLoop() {
	logger.Info("run 过期缓存数据项清理")
	ticker := time.NewTicker(c.gcInterval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-c.stopGc:
			ticker.Stop()
			return
		}
	}
}

// DeleteExpired 删除过期数据项
func (c *UdpClientCache) DeleteExpired() {
	now := time.Now().Unix()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			c.mu.Lock()
			delete(c.items, k)
			c.mu.Unlock()
		}
	}
}

// Set 设置缓存数据项，如果数据项存在则覆盖
func (c *UdpClientCache) Set(udpC *UdpClient, d time.Duration) *UdpClient {
	var e int64
	if d == 0 {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).Unix()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	udpC.Expiration = e
	c.items[udpC.IP] = udpC
	return udpC
}

// Get 获取数据项
func (c *UdpClientCache) Get(k string) (*UdpClient, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.Expired() {
		return nil, false
	}
	return item, true
}

// RetryGet Retry get udp clinet
func (c *UdpClientCache) RetryGet(k string) (*UdpClient, bool) {
	for i := 0; i < 10; i++ {
		c, ok := c.Get(k)
		if ok {
			return c, ok
		}
		time.Sleep(100 * time.Microsecond)
	}
	return nil, false
}

// GetAllKey get all key  当前在线的slave
func (c *UdpClientCache) GetAllKey() []string {
	keys := make([]string, 0)
	for k, _ := range c.items {
		keys = append(keys, k)
	}
	return keys
}

// Handler udp slave 的 Handler
type Handler map[CommandCode]func(ctx *HandlerCtx)

// HandlerCtx udp slave 的 Handler 的上下文
type HandlerCtx struct {
	Cmd    CommandCode
	Conn   *net.UDPConn
	Stream *Stream
}

// Send 发送数据包给 master
func (handler *HandlerCtx) Send(data []byte) error {
	ctxId := ""
	if handler.Stream != nil {
		ctxId = handler.Stream.CtxId
	}
	packate, err := Packet(handler.Cmd, ctxId, data)
	if err != nil {
		return err
	}
	_, err = handler.Conn.Write(packate)
	return err
}

// SendCmd 发送指定指令的数据包给 master
func (handler *HandlerCtx) SendCmd(cmd CommandCode, data []byte) error {
	ctxId := ""
	if handler.Stream != nil {
		ctxId = handler.Stream.CtxId
	}
	packate, err := Packet(cmd, ctxId, data)
	if err != nil {
		return err
	}
	_, err = handler.Conn.Write(packate)
	return err
}

// UDPClient slave udp 客户端启动
func UDPClient(handler Handler, fs []func(conn *net.UDPConn), errFs []func()) {
	logger.Info(conf.SlaveConf.Master)

	sip := net.ParseIP(conf.SlaveConf.Master.Host)
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: sip, Port: conf.SlaveConf.Master.Port}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		logger.Error(err)
	}
	defer conn.Close()

	// 读
	go func() {
		data := make([]byte, 1024*10)
		for {
			n, _, err := conn.ReadFromUDP(data)
			if err != nil {
				logger.Error("error during read: %s", err)
				for _, errf := range errFs {
					go errf()
				}
			}
			//logger.Info(remoteAddr, n)

			if n >= 33 {
				packet := data[:n]
				s := DecryptPacket(packet, n)
				// TODO 中间件
				logger.Info("Command = ", s.Sign, s.CtxId, s.Command.Chinese())
				//logger.Info("读到数据 packet = ", string(packet))
				//logger.Info("s.Command = ", s.Command.Chinese())
				if f, ok := handler[s.Command]; ok {
					//f(s.Command, conn, s)
					go f(&HandlerCtx{
						Conn:   conn,
						Cmd:    s.Command,
						Stream: s,
					})
				} else {
					logger.Info("没有找到命令 ", s.Command)
				}
				//handler[s.Command](conn, s)
			} else {
				logger.Info("传入的数据太小或太大, 建议 25~10240个字节")
			}
		}

	}()

	for _, f := range fs {
		go f(conn)
	}

	select {}
}
