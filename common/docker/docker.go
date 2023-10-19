package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	mountV "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerClient struct {
	Conn *client.Client
	Ctx  context.Context
}

func NewDockerClient() (*DockerClient, error) {
	conn, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return &DockerClient{
		Conn: conn,
		Ctx:  context.Background(),
	}, err
}

// DockerInfo docker 信息
func (d *DockerClient) DockerInfo() (types.Info, error) {
	return d.Conn.Info(d.Ctx)
}

// DockerPs docker 查看所有 容器
func (d *DockerClient) DockerPs() ([]types.Container, error) {
	return d.Conn.ContainerList(context.Background(), types.ContainerListOptions{All: true})
}

// DockerImages docker 查看所有 镜像
func (d *DockerClient) DockerImages() ([]types.ImageSummary, error) {
	return d.Conn.ImageList(context.Background(), types.ImageListOptions{All: false})
}

// DockerPull docker 拉取镜像
func (d *DockerClient) DockerPull(pull entity.DockerPullArg) ([]byte, error) {
	authConfig := types.AuthConfig{
		Username: pull.Name,
		Password: pull.Pass,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		logger.Error("authConfig err : ", err.Error())
		return []byte(""), err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	logger.Info("pull.Image = ", pull.Image, authStr)
	out, err := d.Conn.ImagePull(context.Background(), pull.Image, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		logger.Error("ImagePull err : ", err.Error())
		return []byte(""), err
	}

	bf := make([]byte, 0)
	buf := bytes.NewBuffer(bf)
	defer out.Close()
	io.Copy(buf, out)
	return buf.Bytes(), nil
}

// DockerRun docker 启动一个容器
// TODO : 运行本地镜像
func (d *DockerClient) DockerRun(arg entity.DockerRunArg) (container.ContainerCreateCreatedBody, error) {

	// 端口映射
	portMap := make(nat.PortSet)
	exposedPort := make(nat.PortMap)
	logger.Info("启动一个容器 端口为 : ", arg.PortRelation)
	for k, v := range arg.PortRelation {
		portMap[nat.Port(v)] = struct{}{}
		bind := make([]nat.PortBinding, 0)
		bind = append(bind, nat.PortBinding{
			HostIP:   "0.0.0.0", // 宿主机 ip
			HostPort: k,         // 印射到宿主机的端口
		})
		exposedPort[nat.Port(v)] = bind
	}

	// container.Config
	config := &container.Config{
		Image:     arg.Image, // 镜像名：推荐提前下好
		Tty:       true,      //docker run命令中的-t选项
		OpenStdin: true,      //docker run命令中的-i选项
		//Cmd:        []string{cmd}, //docker 容器中执行的命令
		//WorkingDir: workDir,       //docker容器中的工作目录
		ExposedPorts: portMap,
		Env:          make([]string, 0),
	}

	// Env
	for _, v := range arg.Env {
		if v != "" {
			config.Env = append(config.Env, v)
		}
	}

	// hostConf
	hostConf := &container.HostConfig{}
	hostConf.PortBindings = exposedPort
	if arg.IsAlways {
		hostConf.RestartPolicy = container.RestartPolicy{ // --restart=always
			Name: "always",
		}
	} else {
		hostConf.RestartPolicy = container.RestartPolicy{ // --restart=always
			MaximumRetryCount: 99, // 最大重启99次
		}
	}

	// 挂载Volume
	for _, v := range arg.Volume {
		vList := strings.Split(v, ":")
		if len(vList) != 2 {
			continue
		}
		hostDir := vList[0]
		// 没有目录创建目录
		ok, _ := utils.PathExists(hostDir)
		if !ok {
			_ = utils.Mkdir(hostDir)
		}
		containerDir := vList[1]
		hostConf.Mounts = append(hostConf.Mounts, mountV.Mount{
			Type:   mountV.TypeBind,
			Source: hostDir,
			Target: containerDir,
		})
	}

	// 创建一个container
	resp, err := d.Conn.ContainerCreate(context.Background(),
		config,
		hostConf, // host配置， 基础配置
		nil,      // 网络配置：默认将可以
		nil,      // 平台描述：不用传
		arg.Name, // 容器名：传空会随机分配
	)
	if err != nil {
		logger.Error("ContainerCreate err = ", err)
		return container.ContainerCreateCreatedBody{}, err
	}

	err = d.Conn.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}) // start container
	if err != nil {
		logger.Error("ContainerStart err = ", err)
		return container.ContainerCreateCreatedBody{}, err
	}
	fmt.Println("container start ...")

	// 返回数据  resp
	return resp, nil
}

// DockerStop docker 关闭一个容器
func (d *DockerClient) DockerStop(containerId string) error {
	var t time.Duration = 10 * time.Second
	return d.Conn.ContainerStop(context.Background(), containerId, &t)
}

// DockerRm docker 删除一个容器
func (d *DockerClient) DockerRm(containerId string) error {
	d.DockerStop(containerId)
	return d.Conn.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{})
}

// DockerRmi docker 删除一个镜像
func (d *DockerClient) DockerRmi(imageId string) ([]types.ImageDeleteResponseItem, error) {
	return d.Conn.ImageRemove(context.Background(), imageId, types.ImageRemoveOptions{})
}

// ContainerLog docker 查看容器日志
func (d *DockerClient) ContainerLog(containerId string) (io.ReadCloser, error) {
	return d.Conn.ContainerLogs(context.Background(), containerId, types.ContainerLogsOptions{
		ShowStdout: true,
		Timestamps: true,
		Follow:     true,
		Tail:       "10",
	})
}

// ContainerRename 修改容器名称
func (d *DockerClient) ContainerRename(containerId, newName string) error {
	return d.Conn.ContainerRename(context.Background(), containerId, newName)
}

// DockerState docker state  获取容器的 cpu和mem和网络
func (d *DockerClient) DockerState(containerId string) *entity.ContainerPerformance {
	data, err := d.Conn.ContainerStats(context.Background(), containerId, true)
	if err != nil {
		return nil
	}
	rse1 := ""
	rse2 := ""
	b1 := make([]byte, 1024*4)
	b2 := make([]byte, 1024*4)
	buf1 := make([]byte, 0)
	buf2 := make([]byte, 0)
	data.Body.Read(b1)
	t1 := time.Now().UnixNano()

	for _, v := range b1 {
		if v != '\x00' {
			buf1 = append(buf1, v)
		}
	}
	rse1 = string(buf1)
	time.Sleep(500 * time.Millisecond) // 500ms为一个采样周期

	t2 := time.Now().UnixNano()
	data.Body.Read(b2)
	for _, v := range b2 {
		if v != '\x00' {
			buf2 = append(buf2, v)
		}
	}
	rse2 = string(buf2)
	data.Body.Close()

	var cpu1Num float64 = 0
	cpu1, err := utils.JsonFind(rse1, "/cpu_stats/cpu_usage/total_usage")
	if cpu1 != nil {
		cpu1Num = cpu1.(float64)
	}

	var cpu2Num float64 = 0
	cpu2, err := utils.JsonFind(rse2, "/cpu_stats/cpu_usage/total_usage")
	if cpu2 != nil {
		cpu2Num = cpu2.(float64)
	}

	cpu := ((cpu2Num - cpu1Num) / float64(t2-t1)) * 100
	cpuStr := fmt.Sprintf("%.2f", cpu)
	//logger.Info(cpuStr)

	var memUsageNum float64 = 0
	var memlimitNum float64 = 0
	memUsage, err := utils.JsonFind(rse1, "/memory_stats/usage")
	memlimit, err := utils.JsonFind(rse2, "/memory_stats/limit")
	if memUsage != nil {
		memUsageNum = memUsage.(float64)
	}
	if memlimit != nil {
		memlimit = memlimit.(float64)
	}

	var txNum float64 = 0
	var rxNum float64 = 0
	tx, err := utils.JsonFind(rse1, "/networks/eth0/tx_bytes")
	rx, err := utils.JsonFind(rse1, "/networks/eth0/rx_bytes")
	if tx != nil {
		txNum = tx.(float64)
	}
	if rx != nil {
		rxNum = rx.(float64)
	}

	return &entity.ContainerPerformance{
		ContainerId: containerId,
		Date:        utils.NowTimeStrHMS(),
		CPU:         cpuStr,
		MEM:         fmt.Sprintf("%.4f", (memUsageNum/memlimitNum)*100),
		MEMUsage:    utils.StringValue(int64(memUsageNum)),
		MEMLimit:    utils.StringValue(int64(memlimitNum)),
		Tx:          utils.StringValue(int64(txNum)),
		Rx:          utils.StringValue(int64(rxNum)),
	}
}

// ContainerTop 显示容器中的进程信息。 arguments []string
func (d *DockerClient) ContainerTop(containerId string) (container.ContainerTopOKBody, error) {
	return d.Conn.ContainerTop(context.Background(), containerId, []string{})
}

// ContainerRestart 停止并再次启动容器。
// 它使守护进程等待容器再次启动
// 给定超时的特定时间量。
func (d *DockerClient) ContainerRestart(containerId string) error {
	var t time.Duration = 10 * time.Second
	return d.Conn.ContainerRestart(context.Background(), containerId, &t)
}

// ContainerPause 暂停给定容器的主进程，而不终止它。
func (d *DockerClient) ContainerPause(containerId string) error {
	return d.Conn.ContainerPause(context.Background(), containerId)
}

// GetContainerInfo 获取容器信息
func (d *DockerClient) GetContainerInfo(containerId string) (types.ContainerJSON, error) {
	return d.Conn.ContainerInspect(context.Background(), containerId)
}

// CheckContainerRuning 查看容器是否在运行
func (d *DockerClient) CheckContainerRuning(containerId string) (bool, error) {
	data, err := d.GetContainerInfo(containerId)
	if err != nil {
		return false, err
	}
	return data.State.Running, nil
}

// CheckContainerRuningAt 查看容器是否在运行 如果没有则重启容器
func (d *DockerClient) CheckContainerRuningAt(containerId string) error {
	data, err := d.GetContainerInfo(containerId)
	if err != nil {
		return err
	}
	if !data.State.Running {
		return d.ContainerRestart(containerId)
	}
	return nil
}

// TODO 容器转为镜像

// TODO 查看容器重启次数

// TODO DockerImageOut 将镜像打包输出给master
func (d *DockerClient) DockerImageOut(imageId string) {
	d.Conn.ImageSave(context.Background(), []string{imageId})
}

// testcase docker 容器相关的操作
func (d *DockerClient) testcase(containerId string) {
	//ContainerTach将连接连接到服务器中的容器。
	//它返回一个类型。劫持的连接与被劫持的连接
	//并且读取器接收输出。这取决于被要求关闭的人
	//通过调用types.HijackedResponse.Close可以恢复被劫持的连接。
	//响应上的流格式将采用以下两种格式之一：
	//如果容器使用TTY，则只有一个流（stdout），并且
	//数据直接从容器输出流复制，无需额外的操作
	//多路复用或报头。
	//如果容器*未*使用TTY，则stdout和stderr的流为
	//多路复用。
	//多路复用流的格式如下：
	//[8]字节｛STREAM_TYPE，0，0，SIZE1，SIZE2，SIZE3，SIZE4｝[]字节｝输出｝
	//STREAM_ TYPE可以是1用于标准输出，2用于标准输出
	//SIZE1、SIZE2、SIZE3和SIZE4是uint32的四个字节，编码为big-endian。
	//这是输出的大小。
	//您可以使用github.com/docker/docker/pkg/stdcopy
	d.Conn.ContainerAttach(context.Background(), containerId, types.ContainerAttachOptions{})

	// 将更改应用到容器中并创建新的镜像。
	d.Conn.ContainerCommit(context.Background(), containerId, types.ContainerCommitOptions{})

	// 显示容器文件系统自启动以来的差异。
	d.Conn.ContainerDiff(context.Background(), containerId)

	//将连接附加到服务器中的exec进程。
	//它返回一个类型。劫持的连接与被劫持的连接
	//并且读取器接收输出。这取决于被要求关闭的人
	//通过调用types.HijackedResponse.Close可以恢复被劫持的连接。
	d.Conn.ContainerExecAttach(context.Background(), containerId, types.ExecStartCheck{})

	// 创建新的exec配置以运行exec进程。
	d.Conn.ContainerExecCreate(context.Background(), containerId, types.ExecConfig{})

	// 返回有关docker主机上特定exec进程的信息。
	d.Conn.ContainerExecInspect(context.Background(), containerId)

	//ContainerLogs返回io.ReadCloser中容器生成的日志。
	//由调用方关闭流。
	//响应上的流格式将采用以下两种格式之一：
	//如果容器使用TTY，则只有一个流（stdout），并且
	//数据直接从容器输出流复制，无需额外的操作
	//多路复用或报头。
	//如果容器*未*使用TTY，则stdout和stderr的流为
	//多路复用。
	//多路复用流的格式如下：
	//[8]字节｛STREAM_TYPE，0，0，SIZE1，SIZE2，SIZE3，SIZE4｝[]字节｝输出｝
	//STREAM_ TYPE可以是1用于标准输出，2用于标准输出
	//SIZE1、SIZE2、SIZE3和SIZE4是uint32的四个字节，编码为big-endian。
	//这是输出的大小。
	//您可以使用github.com/docker/docker/pkg/stdcopy
	d.Conn.ContainerLogs(context.Background(), containerId, types.ContainerLogsOptions{})

	// 更改容器内运行的exec进程的tty大小。
	d.Conn.ContainerExecResize(context.Background(), containerId, types.ResizeOptions{})

	// 启动已在docker主机中创建的exec进程。
	d.Conn.ContainerExecStart(context.Background(), containerId, types.ExecStartCheck{})

	// 更改给定容器的名称。
	d.Conn.ContainerRename(context.Background(), containerId, "newContainerName")

	//ContainerWait等待，直到指定的容器处于特定状态
	//由给定条件指示“未运行”（默认），
	//“下一个出口”或“已移除”。
	//如果此客户端的API版本在1.30之前，则忽略该条件，并
	//ContainerWait将立即返回两个通道，作为服务器
	//将等待，就好像条件是“未运行”。
	//如果此客户端的API版本至少为1.30，则ContainerWait将阻止，直到
	//该请求已被服务器确认（具有响应头），
	//然后返回调用者可以等待退出状态的两个通道
	//如果在开始时出现问题，则为容器的错误
	//等待请求或获取响应。这允许调用者：
	//将ContainerWait与其他调用同步，例如指定
	//发出ContainerStart请求之前的“下一个退出”条件。
	//WaitConditionNotRunning WaitCondition = "not-running"
	//WaitConditionNextExit   WaitCondition = "next-exit"
	//WaitConditionRemoved    WaitCondition = "removed"
	d.Conn.ContainerWait(context.Background(), containerId, container.WaitConditionNotRunning)

	// 更新容器的资源
	d.Conn.ContainerUpdate(context.Background(), containerId, container.UpdateConfig{})

	// 显示容器中的进程信息。 arguments []string
	d.Conn.ContainerTop(context.Background(), containerId, []string{})

	//从容器中获取单个stat条目。
	//它与“ContainerStats”的不同之处在于，API不应等待初始化状态
	d.Conn.ContainerStatsOneShot(context.Background(), containerId)

	//返回容器文件系统内路径的统计信息。
	d.Conn.ContainerStatPath(context.Background(), containerId, "")

	// 返回给定容器的近实时统计信息。
	// 由来电者关闭io。ReadCloser返回。  stream
	d.Conn.ContainerStats(context.Background(), containerId, true)

	// 请求守护进程删除未使用的数据
	d.Conn.ContainersPrune(context.Background(), filters.Args{})

	// 返回容器信息及其原始表示形式。
	d.Conn.ContainerInspectWithRaw(context.Background(), containerId, true)

	// 返回容器信息。
	d.Conn.ContainerInspect(context.Background(), containerId)

	// 更改容器的tty大小。
	d.Conn.ContainerResize(context.Background(), containerId, types.ResizeOptions{})

	//检索容器的原始内容
	//并将其作为io.ReadCloser返回。这取决于打电话的人
	//以关闭该流。
	d.Conn.ContainerExport(context.Background(), containerId)
}
