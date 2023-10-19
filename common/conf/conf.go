package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"gitee.com/mangenotework/commander/common/logger"

	yaml "gopkg.in/yaml.v3"
)

var MasterConf *masterConfigs
var SlaveConf *slaveConfigs

// masterConfigs master 配置
type masterConfigs struct {
	HttpServer        *HttpServer `yaml:"httpServer"`
	UdpServer         *UdpServer  `yaml:"udpServer"`
	TcpServer         *TcpServer  `yaml:"tcpServer"`
	ExeStoreHousePath string      `yaml:"exeStoreHousePath"`
	ProjectPath       string      `yaml:"projectPath"`
	Jwt               *Jwt        `yaml:"jwt"`
	DBPath            *DBPath     `yaml:"dbPath"`
}

// slaveConfigs slave 配置
type slaveConfigs struct {
	Master                   *MasterInfo      `yaml:"master"`
	ExeStoreHousePath        string           `yaml:"exeStoreHousePath"`
	MasterHttp               string           `yaml:"masterHttp"`
	ExeStoreHouseLogs        string           `yaml:"exeStoreHouseLogs"`
	ProjectExeStoreHousePath string           `yaml:"projectExeStoreHousePath"`
	DBPath                   *DBPath          `yaml:"dbPath"`
	FileServer               *SlaveFileServer `yaml:"fileServer"`
}

// HttpServer http服务
type HttpServer struct {
	Prod string `yaml:"prod"`
}

// UdpServer udp服务
type UdpServer struct {
	Prod int `yaml:"prod"`
}

// TcpServer tcp服务
type TcpServer struct {
	Prod int `yaml:"prod"`
}

// UdpClient udp客户端
type UdpClient struct {
	Prod string `yaml:"prod"`
}

// MasterInfo master信息
type MasterInfo struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Jwt jwt配置
type Jwt struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"`
}

// DBPath 数据持久化db文件保存路径
type DBPath struct {
	Data        string `yaml:"data"`
	Performance string `yaml:"performance"`
}

// SlaveFileServer Slave 文件服务
type SlaveFileServer struct {
	Port   string `yaml:"port"`
	Secret string `yaml:"secret"`
}

// InitConf 读取yaml文件
// 获取配置
func InitConf(obj interface{}) {
	appConfigPath := "configs.yaml"
	if !fileExists(appConfigPath) {
		panic("【启动失败】 未找到配置文件!")
	}
	logger.Info("[启动]读取配置文件:", appConfigPath)
	//读取yaml文件到缓存中
	config, err := ioutil.ReadFile(appConfigPath)
	if err != nil {
		panic("【启动失败】读取配置文件" + err.Error())
	}
	err = yaml.Unmarshal(config, obj)
	if err != nil {
		panic("【启动失败】读取配置文件" + err.Error())
	}
	b, _ := json.Marshal(obj)
	logger.Info("[conf arg] ", string(b))
}

// MasterInitConf 初始化master配置
func MasterInitConf() {
	InitConf(&MasterConf)
}

// SlaveInitConf 初始化slave配置
func SlaveInitConf() {
	InitConf(&SlaveConf)
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
