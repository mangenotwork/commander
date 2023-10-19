package handler

import (
	"io"
	"log"
	"net"
	"os"
	"time"

	"gitee.com/mangenotework/commander/common/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// wsWrapper
type wsWrapper struct {
	*websocket.Conn
}

func (wsw *wsWrapper) Write(p []byte) (n int, err error) {
	writer, err := wsw.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	//logger.Info("上行数据 : ", string(p))
	return writer.Write(p)
}

func (wsw *wsWrapper) Read(p []byte) (n int, err error) {
	msgType, reader, err := wsw.Conn.NextReader()
	if err != nil {
		logger.Info("Read err = ", err)
		return 0, err
	}
	if msgType != websocket.TextMessage {
		logger.Error("输入不是 TextMessage 类型")
		//continue
	}
	//logger.Info("下行 : ", string(p))
	return reader.Read(p)

}

func WebSocketTerminal(c *gin.Context) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("ws 连接失败 = ", err)
		return
	}
	ip := c.Query("ip")
	port := c.Query("port")
	user := c.Query("user")
	password := c.Query("password")
	go func() {
		logger.Info("新的连接 ")
		rw := io.ReadWriter(&wsWrapper{conn})
		webprintln := func(data string) {
			rw.Write([]byte(data + "\r\n"))
		}
		conn.SetCloseHandler(func(code int, text string) error {
			conn.Close()
			return nil
		})
		go sshHandle(rw, ip, port, user, password, webprintln)
	}()
}

// sshHandle SSH连接
func sshHandle(rw io.ReadWriter, ip, port, user, passwd string, errhandle func(string)) {
	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(passwd)},
		Timeout:         6 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	if port == "" {
		ip = ip + ":22"
	} else {
		ip = ip + ":" + port
	}
	logger.Info("ssh 连接 ")
	client, err := ssh.Dial("tcp", ip, sshConfig)
	if err != nil {
		logger.Error("ssh 连接 失败, err = ", err)
		errhandle(err.Error())
		return
	}

	logger.Info("ssh client = ", client)
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		logger.Error("ssh session 失败, err = ", err)
		errhandle(err.Error())
		return
	}
	defer session.Close()

	//stdinP, err = session.StdinPipe()
	//if err != nil {
	//	return
	//}

	fd := int(os.Stdin.Fd())
	session.Stdout = rw
	session.Stderr = rw
	session.Stdin = rw
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	termWidth, termHeight, err := terminal.GetSize(fd)

	////使用VT100终端来实现tab键提示，上下键查看历史命令，clear键清屏等操作
	////VT100 start
	////windows下不支持VT100
	//oldState, err := terminal.MakeRaw(fd)
	//if err != nil {
	//	log.Fatalln(err.Error())
	//}
	//defer terminal.Restore(fd, oldState)
	////VT100 end

	logger.Info("termWidth, termHeight = ", termWidth, termHeight)
	err = session.RequestPty("xterm", 611, termWidth, modes)
	if err != nil {
		logger.Error("RequestPty err = ", err)
		errhandle(err.Error())
	}

	//启动一个远程shell
	err = session.Shell()
	if err != nil {
		logger.Error("Shell err = ", err)
		errhandle(err.Error())
	}

	//等待远程命令结束或远程shell退出
	err = session.Wait()
	if err != nil {
		logger.Error("Wait err = ", err)
		errhandle(err.Error())
	}
	return
}

func NewTerminal(user string, pass string, addr string) {
	var (
		session *ssh.Session
	)

	c, err := SSHClient(user, pass, addr)
	if err != nil {
		log.Println("ssh 连接失败 : ", err)
	}
	log.Println(c)
	//获取session，这个session是用来远程执行操作的
	if session, err = c.NewSession(); err != nil {
		log.Fatalln("error occurred:", err)
	}
	defer session.Close()
	session.Stdout = os.Stdout // 会话输出关联到系统标准输出设备
	session.Stderr = os.Stderr // 会话错误输出关联到系统标准错误输出设备
	session.Stdin = os.Stdin   // 会话输入关联到系统标准输入设备
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, //output speed = 14.4kbaud
	}
	if err = session.RequestPty("linux", 32, 160, modes); err != nil {
		log.Fatalf("request pty error: %s", err.Error())
	}
	if err = session.Shell(); err != nil {
		log.Fatalf("start shell error: %s", err.Error())
	}
	if err = session.Wait(); err != nil {
		log.Fatalf("return error: %s", err.Error())
	}
}

// SSHClient 连接ssh
// addr : 主机地址, 如: 127.0.0.1:22
// user : 用户
// pass : 密码
// 返回 ssh连接
func SSHClient(user string, pass string, addr string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	sshConn, err := net.Dial("tcp", addr)
	if nil != err {
		return nil, err
	}

	clientConn, chans, reqs, err := ssh.NewClientConn(sshConn, addr, config)
	if nil != err {
		sshConn.Close()
		return nil, err
	}

	client := ssh.NewClient(clientConn, chans, reqs)
	return client, nil
}
