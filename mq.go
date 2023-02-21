package mqclient

type Client struct {
	client client
}

func New(config Config) (*Client, error) {
	cli := new(Client)

	protocol := config.Protocol
	if protocol == "" {
		protocol = ProtocolAMQP091
	}

	client, err := newClientByProtocol(protocol, config)
	if err != nil {
		return nil, err
	}
	cli.client = client

	if config.AutoConnect {
		if err := cli.Connect(); err != nil {
			return nil, err
		}
	}

	return cli, nil
}

func (cli *Client) Connect() error {
	if cli.client.isConnecting() {
		return nil
	}

	return cli.client.connect()
}

func (cli *Client) Close() error {
	return cli.client.close()
}

func (cli *Client) Publish(input PublishInput) error {
	if err := cli.Connect(); err != nil {
		return err
	}

	return cli.client.publish(input)
}
