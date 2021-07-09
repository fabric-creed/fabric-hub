package fabric

import (
	"fmt"
	"github.com/fabric-creed/fabric-sdk-go/pkg/client/channel"
	"github.com/fabric-creed/fabric-sdk-go/pkg/client/ledger"
	"github.com/fabric-creed/fabric-sdk-go/pkg/common/providers/core"
	"github.com/fabric-creed/fabric-sdk-go/pkg/core/config"
	"github.com/fabric-creed/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Client struct {
	ConfigPath     string
	ConfigData     []byte
	Organization   string
	Username       string
	ChannelID      string
	fabricSDK      *fabsdk.FabricSDK
	ledgerManager  map[string]*Ledger
	channelManager map[string]*Channel
}

func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		ledgerManager:  make(map[string]*Ledger, 0),
		channelManager: make(map[string]*Channel, 0),
	}
	for _, opt := range opts {
		opt(c)
	}

	if err := c.setup(); err != nil {
		return nil, fmt.Errorf("failed to setup: %v", err)
	}
	return c, nil
}

func (c *Client) Init() {
	if len(c.ConfigPath) == 0 {
		return
	}
	if err := c.setup(); err != nil {
		logrus.Fatalf("failed to setup: %v", err)
	}
}

func (c *Client) setup() error {
	var configProvider core.ConfigProvider
	if len(c.ConfigData) != 0 {
		configProvider = config.FromRaw(c.ConfigData, "yaml")
	} else {
		configProvider = config.FromFile(c.ConfigPath)
	}
	fabricSDK, err := fabsdk.New(configProvider)
	if err != nil {
		return fmt.Errorf("failed to new fabricSDK: %v", err)
	}
	c.fabricSDK = fabricSDK

	if len(c.ChannelID) != 0 {
		channelProvider := c.fabricSDK.ChannelContext(
			c.ChannelID,
			fabsdk.WithOrg(c.Organization),
			fabsdk.WithUser(c.Username),
		)
		ledgerClient, err := ledger.New(channelProvider)
		if err != nil {
			return errors.Wrap(err, "failed to new ledgerClient")
		}
		c.ledgerManager[c.ChannelID] = &Ledger{client: ledgerClient}
		channelClient, err := channel.New(channelProvider)
		if err != nil {
			return errors.Wrap(err, "failed to new channelClient")
		}
		c.channelManager[c.ChannelID] = &Channel{channelClient}
	}

	return nil
}

func (c *Client) Close() {
	if c.fabricSDK != nil {
		c.fabricSDK.Close()
	}
}

func (c *Client) Channel(channelID string) (*Channel, error) {
	if _, ok := c.channelManager[channelID]; !ok {
		channelProvider := c.fabricSDK.ChannelContext(
			channelID,
			fabsdk.WithOrg(c.Organization),
			fabsdk.WithUser(c.Username),
		)
		channelClient, err := channel.New(channelProvider)
		if err != nil {
			return nil, errors.Wrap(err, "failed to new channel client")
		}
		c.channelManager[channelID] = &Channel{channelClient}
	}

	return c.channelManager[channelID], nil
}

func (c *Client) Ledger(channelID string, isGM bool) (*Ledger, error) {
	if _, ok := c.ledgerManager[channelID]; !ok {
		channelProvider := c.fabricSDK.ChannelContext(
			channelID,
			fabsdk.WithOrg(c.Organization),
			fabsdk.WithUser(c.Username),
		)
		ledgerClient, err := ledger.New(channelProvider)
		if err != nil {
			return nil, errors.Wrap(err, "failed to new ledger client")
		}
		c.ledgerManager[channelID] = &Ledger{client: ledgerClient, isGM: isGM}
	}

	return c.ledgerManager[channelID], nil
}
