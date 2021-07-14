package service

import (
	"github.com/fabric-creed/fabric-hub/config"
	"github.com/fabric-creed/fabric-hub/pkg/client"
	"github.com/fabric-creed/fabric-hub/pkg/common/sw"
	"github.com/fabric-creed/fabric-hub/pkg/fabric"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
)

type HubService struct {
	pb.UnimplementedHubServer

	csp map[string]*sw.SimpleCSP

	channelManager map[string]config.Channel

	fabricManager map[string]*fabric.Client

	hubClientManager map[string]*client.HubClient
}

func NewHubService(options ...Option) *HubService {
	service := &HubService{}
	for _, option := range options {
		option(service)
	}
	return service
}

type Option func(s *HubService)

func WithCSP(csp map[string]*sw.SimpleCSP) Option {
	return func(s *HubService) {
		s.csp = csp
	}
}

func WithChannelManager(channelManager map[string]config.Channel) Option {
	return func(s *HubService) {
		s.channelManager = channelManager
	}
}
func WithFabricManager(fabricManager map[string]*fabric.Client) Option {
	return func(s *HubService) {
		s.fabricManager = fabricManager
	}
}

func WithHubClient(hubClient map[string]*client.HubClient) Option {
	return func(s *HubService) {
		s.hubClientManager = hubClient
	}
}
