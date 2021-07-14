package config

type ViperConfig struct {
	// 存储路径
	DBPath string `json:"dbPath" yaml:"dbPath"`
	// 远端跨链网关
	RemoteFabricNamespace []RemoteFabricNamespace `json:"remoteFabricNamespace" yaml:"remoteFabricNamespace"`
	// 本地跨链网关可负责的fabric
	LocalFabricNamespace []LocalFabricNamespace `json:"localFabricNamespace" yaml:"localFabricNamespace"`
	// 服务配置
	ServerConfig ServerConfig `json:"serverConfig" yaml:"serverConfig"`
}

type RemoteFabricNamespace struct {
	Name         string       `json:"name" yaml:"name"`
	Address      string       `json:"address" yaml:"address"`
	Port         uint32       `json:"port" yaml:"port"`
	ClientConfig ClientConfig `json:"clientConfig" yaml:"clientConfig"`
	Channels     []Channel    `json:"channels" yaml:"channels"`
	CSP          CSP          `json:"csp" yaml:"csp"`
}

type LocalFabricNamespace struct {
	Name string `json:"name" yaml:"name"`
	// fabric config.yaml文件
	FabricConfigPath string
	// sdk 指定的组织名称
	Organization string `json:"organization" yaml:"organization"`
	// sdk 指定的用户
	User string `json:"user" yaml:"user"`
	// 包含的通道
	Channels []Channel
	// 加密
	CSP CSP `json:"csp" yaml:"csp"`
	// 是否为国密
	IsGM bool `json:"isGM" yaml:"isGM"`
}

type Channel struct {
	// 通道名称
	Name string `json:"name" yaml:"name"`
	// 通道ID
	ID string `json:"id" yaml:"id"`
	// 代理合约名称
	ProxyChainCodeName string `json:"proxyChainCodeName" yaml:"proxyChainCodeName"`
	// 路由合约名称
	RouterChainCodeName string `json:"routerChainCodeName" yaml:"routerChainCodeName"`
	// 是否为国密
	IsGM bool
}

type ServerConfig struct {
	// 开启
	UseTLS bool `json:useTLS" yaml:"useTLS"`
	// key 路径
	ServerKeyPath string `json:serverKeyPath" yaml:"serverKeyPath"`
	// cert 路径
	ServerCertPath string `json:serverCertPath" yaml:"serverCertPath"`
	// Ca cert 路径
	ServerRootCAPath string `json:serverRootCAPath" yaml:"serverRootCAPath"`
	// 是否开启双向校验
	RequireClientAuth bool `json:"requireClientAuth" yaml:"requireClientAuth"`
	// client ca 路径
	ClientRootCAPath []string `json:"clientRootCAPath" yaml:"clientRootCAPath"`
	// 端口
	Port int64 `json:"port" yaml:"port"`
}

type ClientConfig struct {
	UseTLS bool `json:useTLS" yaml:"useTLS"`
	// key 路径
	ClientKeyPath string `json:"clientKeyPath" yaml:"clientKeyPath"`
	// cert 路径
	ClientCertPath string `json:"clientCertPath" yaml:"clientCertPath"`
	// cert 路径
	ClientRootCACertPath string `json:"clientRootCACertPath" yaml:"clientRootCACertPath"`
	// Ca cert 路径
	ServerRootCAPath string `json:"serverRootCAPath" yaml:"serverRootCAPath"`
	// 是否是国密
	IsGm bool `json:"isGm" yaml:"isGm"`
}

type CSP struct {
	Cert       string `json:"cert" yaml:"cert"`
	PrivateKey string `json:"privateKey" yaml:"privateKey"`
}
