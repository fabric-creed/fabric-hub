package client

import (
	"github.com/fabric-creed/fabric-hub/pkg/common/grpc"
	"github.com/fabric-creed/fabric-hub/pkg/common/sw"
)

type HubClient struct {
	address string
	port    uint32
	client  *grpc.GRPCClient
	csp     map[string]*sw.SimpleCSP
}

func NewHubClient(address string, port uint32, client *grpc.GRPCClient) (*HubClient, error) {
	return &HubClient{
		address: address,
		port:    port,
		client:  client,
	}, nil
}

func NewGRPCClient(certPath, keyPath, caCertPath, serverCACertPath string, isGm bool) (*grpc.GRPCClient, error) {
	so, err := grpc.ClientSecureOptions(
		certPath,
		keyPath,
		caCertPath,
		serverCACertPath,
		isGm,
	)
	if err != nil {
		return nil, err
	}
	cc := grpc.ClientConfig{
		SecOpts: so,
		KaOpts:  grpc.DefaultKeepaliveOptions,
		Timeout: grpc.DefaultConnectionTimeout,
	}
	return grpc.NewGRPCClient(cc)
}

func (c *HubClient) SetCSP(csp map[string]*sw.SimpleCSP) {
	c.csp = csp
}
