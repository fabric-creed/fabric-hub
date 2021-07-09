package client

import (
	"context"
	"github.com/fabric-creed/fabric-hub/global"
	"github.com/fabric-creed/fabric-hub/pkg/common/grpc"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
	"github.com/pkg/errors"
	"time"
)

type HubClient struct {
	address string
	client  *grpc.GRPCClient
}

type ClientConfig struct {
	UseTLS bool
	// key 路径
	ClientKeyPath string
	// cert 路径
	ClientCertPath string
	// cert 路径
	ClientRootCACertPath string
	// Ca cert 路径
	ServerRootCAPath string
	// 是否是国密
	IsGm bool
}

func NewHubClient(address string, config ClientConfig) (*HubClient, error) {
	secOpts, err := grpc.ClientSecureOptions(
		config.ClientCertPath,
		config.ClientKeyPath,
		config.ClientRootCACertPath,
		config.ServerRootCAPath,
		config.IsGm,
	)
	if err != nil {
		return nil, err
	}
	client, err := grpc.NewGRPCClient(grpc.ClientConfig{
		SecOpts: secOpts,
		KaOpts:  grpc.DefaultKeepaliveOptions,
		Timeout: grpc.DefaultConnectionTimeout,
	})
	if err != nil {
		return nil, err
	}
	return &HubClient{
		address: address,
		client:  client,
	}, nil
}

