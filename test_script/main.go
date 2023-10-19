package main

import (
	"context"
	gt "github.com/mangenotwork/gathertool"
	"io/ioutil"
	"log"
	"os/exec"
)

func main(){
	gt.CPUMax()

	// 解决端口不够用的问题，connect: cannot assign requested address
	// 调低端口释放后的等待时间，默认为60s，修改为15~30s：
	LinuxSendCommand("sysctl -w net.ipv4.tcp_fin_timeout=22")
	// 修改tcp/ip协议配置， 通过配置/proc/sys/net/ipv4/tcp_tw_resue, 默认为0，修改为1，释放TIME_WAIT端口给新连接使用：
	LinuxSendCommand("sysctl -w net.ipv4.tcp_timestamps=1")
	// 修改tcp/ip协议配置，快速回收socket资源，默认为0，修改为1：
	LinuxSendCommand("sysctl -w net.ipv4.tcp_tw_recycle=1")
	// 允许端口重用：
	LinuxSendCommand("sysctl -w net.ipv4.tcp_tw_reuse = 1")
	// 全连接队列有空位时，接受到客户端的重试ACK，任然会触发服务端连接成功。
	LinuxSendCommand("sysctl -w net.ipv4.tcp_abort_on_overflow = 0")

	// 普通 GET api压测
	url := "http://192.168.0.9:13335/"
	//url := "http://192.168.0.9:50009/"
	// 请求10000次 并发数 1000
	test := gt.NewTestUrl(url,"Get",10000,1000)
	test.Run()
	//test.Run(gt.SucceedFunc(func(ctx *gt.Context){
	//	log.Println(ctx.JobNumber, "测试完成!!", ctx.Ms)
	//}))
}



func LinuxSendCommand(command string) (opStr string) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx,"/bin/bash", "-c", command)
	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		log.Fatal("ERR stdout : ", stdoutErr)
	}
	defer cancel()
	defer stdout.Close()
	if startErr := cmd.Start(); startErr != nil {
		log.Fatal("ERR Start : ", startErr)
	}
	opBytes, opBytesErr := ioutil.ReadAll(stdout)
	if opBytesErr != nil {
		//log.Println(string(opBytes))
		opStr = ""
	}
	opStr = string(opBytes)
	//log.Println(opStr)
	cmd.Wait()
	return
}

