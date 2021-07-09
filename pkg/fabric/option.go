package fabric

type Option func(*Client)

func WithConfigPath(configPath string) Option {
	return func(c *Client) {
		c.ConfigPath = configPath
	}
}

func WithOrganization(organization string) Option {
	return func(c *Client) {
		c.Organization = organization
	}
}

func WithUsername(username string) Option {
	return func(c *Client) {
		c.Username = username
	}
}

func WithChannelID(channelID string) Option {
	return func(c *Client) {
		c.ChannelID = channelID
	}
}
