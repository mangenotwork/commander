package utils

import (
	"gitee.com/mangenotework/commander/common/logger"
	"net"
)

// Tcper tcp客户端
func Tcper(ip string) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		logger.Error(err)
		return false
	}
	_, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logger.Error(err)
		return false
	}
	return true
}
