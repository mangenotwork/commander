package entity

// GatewayBase 网关基础参数
type GatewayBase struct {
	Slave string // 部署在哪个主机上
	Ports []string // 8080:80
	ProjectName string  // 绑定到的项目
	LVS string // 负载均衡层级   L4  L7
	LVSModel string // 负载均衡模式   random  poll
	IsClose string // 是否关闭   1:关闭
	Create string // 创建时间
}

// GatewayArg 网关参数
type GatewayArg struct {
	Ports []string // 8080:80
	ProjectName string  // 绑定到的项目
	LVS string // 负载均衡层级   L4  L7
	LVSModel string // 负载均衡模式   random  poll
}
