package mqclient

import (
	"sync"

	"github.com/ghosind/utils"
	"github.com/rabbitmq/amqp091-go"
)

type amqp091Client struct {
	uris     []string
	conn     *amqp091.Connection
	connChan chan *amqp091.Error
	channel  *amqp091.Channel
	chanChan chan *amqp091.Error

	isConnected bool
	connMutex   sync.Mutex

	autoReconnect bool
}

func newAmqp091Client(config Config) (*amqp091Client, error) {
	cli := new(amqp091Client)

	cli.parseServers(config.Servers)

	cli.isConnected = false

	cli.autoReconnect = config.AutoReconnect

	return cli, nil
}

func (cli *amqp091Client) connect() error {
	cli.connMutex.Lock()
	defer cli.connMutex.Unlock()

	if cli.conn == nil || cli.conn.IsClosed() {
		for _, uri := range cli.uris {
			conn, err := amqp091.Dial(uri)
			if err != nil {
				continue
			}

			cli.conn = conn

			cli.setConnCloseListener()
		}

		if cli.conn == nil {
			return ErrNoAvailableServer
		}
	}

	if cli.channel == nil || cli.channel.IsClosed() {
		channel, err := cli.conn.Channel()
		if err != nil {
			return err
		}

		cli.channel = channel

		cli.setChanCloseListener()
	}

	cli.isConnected = true

	return nil
}

func (cli *amqp091Client) close() error {
	cli.connMutex.Lock()
	defer cli.connMutex.Unlock()

	if cli.channel != nil && !cli.channel.IsClosed() {
		if err := cli.channel.Close(); err != nil {
			return err
		}

		cli.channel = nil
	}

	if cli.conn != nil && !cli.conn.IsClosed() {
		if err := cli.conn.Close(); err != nil {
			return err
		}

		cli.conn = nil
	}

	cli.isConnected = false

	return nil
}

func (cli *amqp091Client) isConnecting() bool {
	if !cli.isConnected {
		return false
	}

	if cli.channel == nil || cli.channel.IsClosed() {
		return false
	}

	if cli.conn == nil || cli.conn.IsClosed() {
		return false
	}

	return true
}

func (cli *amqp091Client) parseServers(servers []ServerConfig) {
	if len(servers) == 0 {
		cli.uris = make([]string, 1)
		cli.uris[0] = amqp091.URI{
			Scheme:   "amqp",
			Host:     "127.0.0.1",
			Port:     amqpDefaultPort,
			Username: amqpDefaultUser,
			Password: amqpDefaultPass,
			Vhost:    amqpDefaultVHost,
		}.String()
		return
	}

	cli.uris = make([]string, 0, len(servers))
	for _, server := range servers {
		uri := amqp091.URI{
			Scheme: utils.Conditional(server.SSL, "amqps", "amqp"),
			Host:   utils.Conditional(server.Host == "", "127.0.0.1", server.Host),
			Port: utils.Conditional(
				server.Port == 0,
				utils.Conditional(server.SSL, amqpDefaultPortSSL, amqpDefaultPort),
				server.Port,
			),
			Username: utils.Conditional(server.User == "", amqpDefaultUser, server.User),
			Password: utils.Conditional(server.Pass == "", amqpDefaultPass, server.Pass),
			Vhost:    utils.Conditional(server.VHost == "", amqpDefaultVHost, server.VHost),
		}

		cli.uris = append(cli.uris, uri.String())
	}
}

func (cli *amqp091Client) setConnCloseListener() {
	cli.connChan = make(chan *amqp091.Error)
	cli.conn.NotifyClose(cli.connChan)

	go func() {
		for {
			err := <-cli.connChan
			if err != nil {
				if cli.autoReconnect {
					// TODO: reconnect
				} else {
					cli.connMutex.Lock()
					defer cli.connMutex.Unlock()

					cli.conn = nil
					cli.channel = nil
					cli.isConnected = false
				}

				break
			}
		}
	}()
}

func (cli *amqp091Client) setChanCloseListener() {
	cli.chanChan = make(chan *amqp091.Error)
	cli.channel.NotifyClose(cli.chanChan)

	go func() {
		for {
			err := <-cli.chanChan
			if err != nil {
				if cli.autoReconnect {
					// TODO: reconnect
				} else {
					cli.connMutex.Lock()
					defer cli.connMutex.Unlock()

					cli.channel = nil
					// Maybe the channel was closed only, so, we just set the client's channel to nil.
				}

				break
			}
		}
	}()
}
