package fabric

import (
	"github.com/fabric-creed/fabric-sdk-go/pkg/client/channel"
)

type Channel struct {
	client *channel.Client
}

func (c *Channel) ChannelExecute(
	request channel.Request,
	options ...channel.RequestOption,
) (channel.Response, error) {
	return c.client.Execute(request, options...)
}

func (c *Channel) ChannelQuery(
	request channel.Request,
	options ...channel.RequestOption,
) (channel.Response, error) {
	return c.client.Query(request, options...)
}
