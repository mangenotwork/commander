package cmd

import (
	"context"
	"io/ioutil"
	"os/exec"

	"gitee.com/mangenotework/commander/common/logger"
)

// LinuxSendCommand Linux Send Command Linux执行命令
// command 要执行的命令
func LinuxSendCommand(command string) (opStr string) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx,"/bin/bash", "-c", command)
	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		logger.Error("ERR stdout : ", stdoutErr)
	}
	defer cancel()
	defer stdout.Close()
	if startErr := cmd.Start(); startErr != nil {
		logger.Error("ERR Start : ", startErr)
	}
	opBytes, opBytesErr := ioutil.ReadAll(stdout)
	if opBytesErr != nil {
		opStr = opBytesErr.Error()
	}
	opStr = string(opBytes)
	cmd.Wait()
	return
}

// WindowsSendCommand Windows Send Command Linux执行命令
// command 要执行的命令
func WindowsSendCommand(command []string) (opStr string) {
	ctx, cancel := context.WithCancel(context.Background())
	if len(command) < 1 {
		return ""
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:len(command)]...)
	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		logger.Error("ERR stdout : ", stdoutErr)
	}
	defer cancel()
	defer stdout.Close()
	if startErr := cmd.Start(); startErr != nil {
		logger.Error("ERR Start : ", startErr)
	}
	opBytes, opBytesErr := ioutil.ReadAll(stdout)
	if opBytesErr != nil {
		logger.Error(opBytesErr)
		opStr = ""
	}
	opStr = string(opBytes)
	cmd.Wait()
	return
}

// TODO WindwsSendPipe 执行windows 管道命令
func WindwsSendPipe(command1, command2 []string) (opStr string) {
	return ""
}
