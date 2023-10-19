package entity

import (
	"github.com/docker/docker/api/types/container"
)

// DockerPullArg docker pull
type DockerPullArg struct {
	Name   string
	Pass   string
	Image  string
	TaskId string
}

// DockerPullResult docker pull
type DockerPullResult struct {
	Data   string
	Err    error
	TaskId string
}

// DockerRunArg docker run
type DockerRunArg struct {
	Image        string
	ImageUser    string
	ImagePass    string
	RandomPort   int               // 0 false  1 true  是否随机端口
	PortRelation map[string]string // 端口映射
	Name         string
	Env          []string
	Volume       []string
	TaskId       string
	IsProject    bool // 是否属于项目
	Project      string
	IsAlways     bool // 是否 always
}

// DockerImageRunArg docker image run
type DockerImageRunArg struct {
	ImageId      string
	RandomPort   int // 0 false  1 true
	PortRelation map[string]string
	Name         string
	Env          []string
	TaskId       string
	IsProject    bool
	Project      string
	IsAlways     bool
}

// DockerImageRunRse 结果
type DockerImageRunRse struct {
	TaskId string
	Rse    string
}

// UpdateDockerImageArg docker容器项目更新镜像传入参数
type UpdateDockerImageArg struct {
	Arg          DockerRunArg
	ContainerIds []string
}

// DockerRunResult docker run
type DockerRunResult struct {
	ID        string
	Warnings  []string
	Err       error
	TaskId    string
	IsProject bool
	Project   string
	Port      map[string]string
}

// ContainerTopResult container top
type ContainerTopResult struct {
	Data container.ContainerTopOKBody
	Err  error
}

// ContainerReNameArg container rename
type ContainerReNameArg struct {
	ContainerId string
	NewName     string
}

// ContainerPerformance container performance
type ContainerPerformance struct {
	ContainerId string
	Date        string
	CPU         string // 使用率
	MEM         string // 使用率
	MEMUsage    string
	MEMLimit    string
	Tx          string // 网络tx
	Rx          string // 网络rx
}

// DockerContainerDeploy docker container deploy
type DockerContainerDeploy struct {
	Slave       string
	Project     string
	ContainerId string
	TaskID      string
	Port        map[string]string
}

// DockerContainerRmArg docker container rm
type DockerContainerRmArg struct {
	IsProject   string // 1:是项目   0:不是
	Slave       string
	Project     string
	ContainerId string
	TaskID      string
}

// DockerContainerRmRse docker container rm
type DockerContainerRmRse struct {
	IsProject   string // 1:是项目   0:不是
	Slave       string
	Project     string
	ContainerId string
	TaskId      string
	Rse         string
}

// DockerStopArg docker stop arg
type DockerStopArg struct {
	TaskId      string
	ContainerId string
}

// DockerStopRse docker stop rse
type DockerStopRse struct {
	TaskId      string
	Rse         string
	ContainerId string
}

// DockerRmiArg docker rmi arg
type DockerRmiArg struct {
	TaskId  string
	ImageId string
}

// DockerRmiRse docker rmi rse
type DockerRmiRse struct {
	TaskId  string
	ImageId string
	Rse     string
}

// DockerRestartArg docker restart arg
type DockerRestartArg struct {
	TaskId      string
	ContainerId string
}

// DockerRestartRse docker restart rse
type DockerRestartRse struct {
	TaskId      string
	ContainerId string
	Rse         string
}

// DockerPauseArg docker pause arg
type DockerPauseArg struct {
	TaskId      string
	ContainerId string
}

// DockerPauseRse docker pause rse
type DockerPauseRse struct {
	TaskId      string
	ContainerId string
	Rse         string
}
