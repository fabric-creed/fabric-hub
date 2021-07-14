package main

import (
	"fmt"
	"github.com/fabric-creed/fabric-hub/global"
	"github.com/fabric-creed/fabric-hub/pkg/adopter"
	"github.com/fabric-creed/fabric-hub/pkg/adopter/fabric"
	cgrpc "github.com/fabric-creed/fabric-hub/pkg/common/grpc"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
	"github.com/fabric-creed/fabric-hub/pkg/service"
	"github.com/fabric-creed/grpc/reflection"
	"log"
)

func main() {
	so, err := cgrpc.ServerSecureOptions(
		global.Config.GRPCServerConfig.UseTLS,
		global.Config.GRPCServerConfig.ServerCertPath,
		global.Config.GRPCServerConfig.ServerKeyPath,
		global.Config.GRPCServerConfig.ServerCertPath,
		global.Config.GRPCServerConfig.RequireClientAuth,
		global.Config.GRPCServerConfig.ClientRootCAPath,
	)
	if err != nil {
		panic(err)
	}

	grpcServer, err := cgrpc.NewGRPCServer(fmt.Sprintf(":%d", global.Config.GRPCServerConfig.Port),
		cgrpc.ServerConfig{
			SecOpts:           so,
			ConnectionTimeout: cgrpc.DefaultConnectionTimeout,
			KaOpts:            cgrpc.DefaultKeepaliveOptions,
		})
	if err != nil {
		panic(err)
	}

	pb.RegisterHubServer(grpcServer.Server(), service.NewHubService(
		service.WithCSP(global.Config.CSPManager),
		service.WithChannelManager(global.Config.LocalChannelManager),
		service.WithFabricManager(global.Config.FabricClientManager),
		service.WithHubClient(global.Config.HubClientManager),
	))
	reflection.Register(grpcServer.Server())

	log.Printf("grpc server is starting, listen on %d \n", global.Config.GRPCServerConfig.Port)

	for id, channel := range global.Config.LocalChannelManager {
		fab := global.Config.FabricClientManager[id]
		fabAdopter := fabric.NewFabric(global.Config.DBPath, channel.Name, channel.RouterChainCodeName, channel.IsGM,
			fabric.WithFabricClient(fab),
			fabric.WithHubClientMap(global.Config.HubClientManager))

		go adopter.NewCrossChainTask(fabAdopter).Run()
	}

	if err := grpcServer.Start(); err != nil {
		panic(err)
	}
}
