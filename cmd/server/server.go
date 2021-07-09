package main

import (
	"fmt"
	"github.com/fabric-creed/fabric-hub/global"
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
			ConnectionTimeout: cgrpc.DefaultConnectionTimeout,
			SecOpts:           so,
			KaOpts:            cgrpc.DefaultKeepaliveOptions,
		})
	if err != nil {
		panic(err)
	}

	pb.RegisterHubServer(grpcServer.Server(), &service.HubService{})
	reflection.Register(grpcServer.Server())

	log.Printf("grpc server is starting, listen on %d \n", global.Config.GRPCServerConfig.Port)
	if err := grpcServer.Start(); err != nil {
		panic(err)
	}
}
