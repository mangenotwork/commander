package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"gitee.com/mangenotework/commander/common/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
	"sync"
	"time"
)

func SSH2(c *gin.Context) {
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("连接失败: ", err)
		return
	}
	defer wsConn.Close()

	client, err := NewSshClient()
	if err != nil {
		logger.Error("ssh连接失败: ", err)
		return
	}
	defer client.Close()

	ssConn, err := NewSshConn(120, 32, client)
	if err != nil {
		return
	}
	defer ssConn.Close()

	quitChan := make(chan bool, 3)

	var logBuff = new(bytes.Buffer)

	// most messages are ssh output, not webSocket input
	go ssConn.ReceiveWsMsg(wsConn, logBuff, quitChan)
	go ssConn.SendComboOutput(wsConn, quitChan)
	go ssConn.SessionWait(quitChan)

	<-quitChan

	//sws, err := NewLogicSshWsSession(120, 32, true, client, wsConn)
	//if err != nil {
	//	logger.Error("ssh ws session 错误")
	//	return
	//}
	//defer sws.Close()
	//quitChan := make(chan bool, 3)
	//sws.Start(quitChan)
	//go sws.Wait(quitChan)
	//<-quitChan
}

func NewSshClient() (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Second * 5,
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
		Auth: []ssh.AuthMethod{ssh.Password("123")},
	}

	addr := fmt.Sprintf("%s:%d", "192.168.3.245", 22)
	c, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewSshConn(cols, rows int, sshClient *ssh.Client) (*SshConn, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	// we set stdin, then we can write data to ssh server via this stdin.
	// but, as for reading data from ssh server, we can set Session.Stdout and Session.Stderr
	// to receive data from ssh server, and write back to somewhere.
	stdinP, err := sshSession.StdinPipe()
	if err != nil {
		return nil, err
	}

	comboWriter := new(wsBufferWriter)
	//ssh.stdout and stderr will write output into comboWriter
	sshSession.Stdout = comboWriter
	sshSession.Stderr = comboWriter

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := sshSession.RequestPty("xterm", rows, cols, modes); err != nil {
		return nil, err
	}
	// Start remote shell
	if err := sshSession.Shell(); err != nil {
		return nil, err
	}
	return &SshConn{StdinPipe: stdinP, ComboOutput: comboWriter, Session: sshSession}, nil
}

// write data to WebSocket
// the data comes from ssh server.
type wsBufferWriter struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

// implement Write interface to write bytes from ssh server into bytes.Buffer.
func (w *wsBufferWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}

// SshConn connect to ssh server using ssh session.
type SshConn struct {
	// calling Write() to write data into ssh server
	StdinPipe io.WriteCloser
	// Write() be called to receive data from ssh server
	ComboOutput *wsBufferWriter
	Session     *ssh.Session
}

func (s *SshConn) Close() {
	if s.Session != nil {
		s.Session.Close()
	}

}

type wsMsg struct {
	Type string `json:"type"`
	Cmd  string `json:"cmd"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

const (
	wsMsgCmd    = "cmd"
	wsMsgResize = "resize"
)

// ReceiveWsMsg  receive websocket msg do some handling then write into ssh.session.stdin
func (ssConn *SshConn) ReceiveWsMsg(wsConn *websocket.Conn, logBuff *bytes.Buffer, exitCh chan bool) {
	//tells other go routine quit
	defer setQuit(exitCh)
	for {
		select {
		case <-exitCh:
			return
		default:
			//read websocket msg
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				logger.Error("reading webSocket message failed")
				return
			}
			//unmashal bytes into struct
			msgObj := wsMsg{
				Type: "cmd",
				Cmd:  "",
				Rows: 50,
				Cols: 180,
			}
			//if err := json.Unmarshal(wsData, &msgObj); err != nil {
			//  logger.Error("unmarshal websocket message failed")
			//}
			switch msgObj.Type {
			case wsMsgResize:
				//handle xterm.js size change
				if msgObj.Cols > 0 && msgObj.Rows > 0 {
					if err := ssConn.Session.WindowChange(msgObj.Rows, msgObj.Cols); err != nil {
						logger.Error("ssh pty change windows size failed")
					}
				}
			case wsMsgCmd:
				//handle xterm.js stdin
				//decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Cmd)
				decodeBytes := wsData
				if err != nil {
					logger.Error("websock cmd string base64 decoding failed")
				}
				if _, err := ssConn.StdinPipe.Write(decodeBytes); err != nil {
					logger.Error("ws cmd bytes write to ssh.stdin pipe failed")
				}
				//write input cmd to log buffer
				if _, err := logBuff.Write(decodeBytes); err != nil {
					logger.Error("write received cmd into log buffer failed")
				}
			}
		}
	}
}

func (ssConn *SshConn) SendComboOutput(wsConn *websocket.Conn, exitCh chan bool) {
	//tells other go routine quit
	defer setQuit(exitCh)

	//every 120ms write combine output bytes into websocket response
	tick := time.NewTicker(time.Millisecond * time.Duration(120))
	//for range time.Tick(120 * time.Millisecond){}
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			//write combine output bytes into websocket response
			if err := flushComboOutput(ssConn.ComboOutput, wsConn); err != nil {
				logger.Error("ssh sending combo output to webSocket failed")
				return
			}
		case <-exitCh:
			return
		}
	}
}

// flushComboOutput flush ssh.session combine output into websocket response
func flushComboOutput(w *wsBufferWriter, wsConn *websocket.Conn) error {
	if w.buffer.Len() != 0 {
		err := wsConn.WriteMessage(websocket.TextMessage, w.buffer.Bytes())
		if err != nil {
			return err
		}
		w.buffer.Reset()
	}
	return nil
}

func (ssConn *SshConn) SessionWait(quitChan chan bool) {
	if err := ssConn.Session.Wait(); err != nil {
		logger.Error("ssh session wait failed")
		setQuit(quitChan)
	}
}

func NewLogicSshWsSession(cols, rows int, isAdmin bool, sshClient *ssh.Client, wsConn *websocket.Conn) (*LogicSshWsSession, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	stdinP, err := sshSession.StdinPipe()
	if err != nil {
		return nil, err
	}

	comboWriter := new(safeBuffer)
	logBuf := new(safeBuffer)
	inputBuf := new(safeBuffer)
	//ssh.stdout and stderr will write output into comboWriter
	sshSession.Stdout = comboWriter
	sshSession.Stderr = comboWriter

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := sshSession.RequestPty("xterm", rows, cols, modes); err != nil {
		return nil, err
	}
	// Start remote shell
	if err := sshSession.Shell(); err != nil {
		return nil, err
	}
	return &LogicSshWsSession{
		stdinPipe:       stdinP,
		comboOutput:     comboWriter,
		logBuff:         logBuf,
		inputFilterBuff: inputBuf,
		session:         sshSession,
		wsConn:          wsConn,
		isAdmin:         isAdmin,
		IsFlagged:       false,
	}, nil
}

type safeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *safeBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}
func (w *safeBuffer) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Bytes()
}
func (w *safeBuffer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}

type LogicSshWsSession struct {
	stdinPipe       io.WriteCloser
	comboOutput     *safeBuffer //ssh 终端混合输出
	logBuff         *safeBuffer //保存session的日志
	inputFilterBuff *safeBuffer //用来过滤输入的命令和ssh_filter配置对比的
	session         *ssh.Session
	wsConn          *websocket.Conn
	isAdmin         bool
	IsFlagged       bool `comment:"当前session是否包含禁止命令"`
}

// Close 关闭
func (sws *LogicSshWsSession) Close() {
	if sws.session != nil {
		sws.session.Close()
	}
	if sws.logBuff != nil {
		sws.logBuff = nil
	}
	if sws.comboOutput != nil {
		sws.comboOutput = nil
	}
}

func (sws *LogicSshWsSession) Start(quitChan chan bool) {
	go sws.receiveWsMsg(quitChan)
	go sws.sendComboOutput(quitChan)
}

func setQuit(ch chan bool) {
	ch <- true
}

// receiveWsMsg  receive websocket msg do some handling then write into ssh.session.stdin
func (sws *LogicSshWsSession) receiveWsMsg(exitCh chan bool) {
	wsConn := sws.wsConn
	//tells other go routine quit
	defer setQuit(exitCh)
	for {
		select {
		case <-exitCh:
			return
		default:
			//read websocket msg
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				logger.Error("read websocket msg err=  ", err)
			}

			d := string(wsData)

			logger.Info(d)

			decodeBytes, err := base64.StdEncoding.DecodeString(d)
			if err != nil {
				logger.Error("base64.StdEncoding.DecodeString err = ", err)
			}
			logger.Info("decodeBytes = ", decodeBytes)

			sws.sendWebsocketInputCommandToSshSessionStdinPipe(wsData)
		}
	}
}

// sendWebsocketInputCommandToSshSessionStdinPipe
func (sws *LogicSshWsSession) sendWebsocketInputCommandToSshSessionStdinPipe(cmdBytes []byte) {
	if _, err := sws.stdinPipe.Write(cmdBytes); err != nil {
		logger.Error("ws cmd bytes write to ssh.stdin pipe failed")
	}
}

func (sws *LogicSshWsSession) sendComboOutput(exitCh chan bool) {
	wsConn := sws.wsConn
	//todo 优化成一个方法
	//tells other go routine quit
	defer setQuit(exitCh)

	//every 120ms write combine output bytes into websocket response
	tick := time.NewTicker(time.Millisecond * time.Duration(60))
	//for range time.Tick(120 * time.Millisecond){}
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if sws.comboOutput == nil {
				return
			}
			bs := sws.comboOutput.Bytes()
			if len(bs) > 0 {
				err := wsConn.WriteMessage(websocket.TextMessage, bs)
				if err != nil {
					logger.Error("ssh sending combo output to webSocket failed")
				}
				_, err = sws.logBuff.Write(bs)
				if err != nil {
					logger.Error("combo output to log buffer failed")
				}
				sws.comboOutput.buffer.Reset()
			}

		case <-exitCh:
			return
		}
	}
}

func (sws *LogicSshWsSession) Wait(quitChan chan bool) {
	if err := sws.session.Wait(); err != nil {
		logger.Error("ssh session wait failed")
		setQuit(quitChan)
	}
}

func (sws *LogicSshWsSession) LogString() string {
	return sws.logBuff.buffer.String()
}
