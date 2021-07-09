package global

import (
	"fmt"
	"github.com/fabric-creed/fabric-hub/pkg/common/grpc"
	"github.com/fabric-creed/fabric-hub/pkg/common/sw"
	"github.com/fabric-creed/fabric-hub/pkg/fabric"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"log"
)

var Config = Configuration{
	LocalChannelManager: make(map[string]Channel, 0),
	GRPCClientManager:   make(map[string]*grpc.GRPCClient, 0),
	CSPManager:          make(map[string]*sw.SimpleCSP, 0),
	FabricClientManager: make(map[string]*fabric.Client, 0),
}

type Configuration struct {
	// bbolt 存储路径
	DBPath string
	// fabric客户端管理器
	FabricClientManager map[string]*fabric.Client
	// grpc客户端管理器
	GRPCClientManager map[string]*grpc.GRPCClient
	// 本fabric环境的通道信息
	LocalChannelManager map[string]Channel
	// grpc server配置
	GRPCServerConfig ServerConfig
	// 各链的公私钥对,用于签名和验牵
	CSPManager map[string]*sw.SimpleCSP
}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	vc := ViperConfig{}
	err = viper.Unmarshal(&vc)
	if err != nil {
		log.Fatal(err)
	}
	Config.DBPath = vc.DBPath
	if vc.DBPath == "" {
		Config.DBPath = "./store"
	}

	parseRemoteNamespaceConfig(vc.RemoteFabricNamespace)

	parseLocalNamespaceConfig(vc.LocalFabricNamespace)

	Config.GRPCServerConfig = vc.ServerConfig
}

func parseRemoteNamespaceConfig(namespaces []RemoteFabricNamespace) {
	// 远端网关不要重名，channelID不能为空
	var nameMap = make(map[string]string, 0)
	for _, namespace := range namespaces {
		if _, ok := nameMap[namespace.Name]; ok {
			panic(fmt.Errorf("remote namespace %s is existed", namespace.Name))
		}

		// 一个namespace下创建一个grpcClient即可
		client, err := newGRPCClient(namespace.ClientConfig)
		if err != nil {
			panic(errors.Wrapf(err, "failed to create grpc client:%s", namespace.ClientConfig))
		}

		// 一个namespace需要一个csp,cert必填
		if namespace.CSP.Cert == "" {
			panic(errors.New(fmt.Sprintf("the cert in  %s namespace is empty", namespace.Name)))
		}
		ks, err := sw.NewSimpleCSP(namespace.CSP.PrivateKey, namespace.CSP.Cert)
		if err != nil {
			panic(errors.Wrapf(err, "failed to new key store in % namespace", namespace.Name))
		}

		for _, channel := range namespace.Channels {
			if channel.ID == "" {
				panic(fmt.Errorf("the channel id is empty in remote namespace %s", namespace.Name))
			}
			Config.GRPCClientManager[channel.ID] = client
			Config.CSPManager[channel.ID] = ks
		}

		nameMap[namespace.Name] = namespace.Name
	}
}

func parseLocalNamespaceConfig(namespaces []LocalFabricNamespace) {
	// 本地的fabric不要重名，channel唯一表示
	var nameMap = make(map[string]string, 0)
	for _, namespace := range namespaces {
		if _, ok := nameMap[namespace.Name]; ok {
			panic(fmt.Errorf("local namespace %s is existed", namespace.Name))
		}
		// 读取配置文件，并实例化client
		client, err := fabric.NewClient(
			fabric.WithConfigPath(namespace.FabricConfigPath),
			fabric.WithOrganization(namespace.Organization),
			fabric.WithUsername(namespace.User),
		)
		if err != nil {
			panic(errors.Wrapf(err, "failed to create fabric client, namespace:%s, config path: %s",
				namespace.Name, namespace.FabricConfigPath))
		}

		if namespace.CSP.PrivateKey == "" {
			panic(errors.New(fmt.Sprintf("the key in  %s namespace is empty", namespace.Name)))
		}
		ks, err := sw.NewSimpleCSP(namespace.CSP.PrivateKey, namespace.CSP.Cert)
		if err != nil {
			panic(errors.Wrapf(err, "failed to new key store in % namespace", namespace.Name))
		}

		for _, channel := range namespace.Channels {
			if channel.ID == "" {
				panic(fmt.Errorf("the channel id is empty in local namespace %s", namespace.Name))
			}
			Config.FabricClientManager[channel.ID] = client
			Config.CSPManager[channel.ID] = ks
			Config.LocalChannelManager[channel.ID] = channel
		}

		nameMap[namespace.Name] = namespace.Name
	}
}

func newGRPCClient(config ClientConfig) (*grpc.GRPCClient, error) {
	so, err := grpc.ClientSecureOptions(
		config.ClientCertPath,
		config.ClientKeyPath,
		config.ClientRootCACertPath,
		config.ServerRootCAPath,
		config.IsGm,
	)
	if err != nil {
		panic(err)
	}
	cc := grpc.ClientConfig{
		SecOpts: so,
		KaOpts:  grpc.DefaultKeepaliveOptions,
		Timeout: grpc.DefaultConnectionTimeout,
	}
	return grpc.NewGRPCClient(cc)
}
