package mqclient

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/ghosind/utils"
	"github.com/go-stomp/stomp"
	"github.com/go-stomp/stomp/frame"
)

const (
	stompDefaultPort    int    = 61613
	stompDefaultPortSSL int    = 61614
	stompDefaultUser    string = "guest"
	stompDefaultPass    string = "guest"
)

type stompServer struct {
	address string
	ssl     bool
}

type stompClient struct {
	servers []*stompServer
	user    string
	pass    string

	conn    *stomp.Conn
	netConn io.ReadWriteCloser

	isConnected bool
	connMutex   sync.Mutex

	autoReconnect bool
}

func newStompClient(config Config) (*stompClient, error) {
	cli := new(stompClient)

	cli.user = utils.Conditional(config.User == "", stompDefaultUser, config.User)
	cli.pass = utils.Conditional(config.Pass == "", stompDefaultPass, config.Pass)
	cli.autoReconnect = config.AutoReconnect

	cli.parseServers(config)

	cli.isConnected = false

	return cli, nil
}

func (cli *stompClient) connect() error {
	cli.connMutex.Lock()
	defer cli.connMutex.Unlock()

	if cli.netConn == nil {
		for _, server := range cli.servers {
			if netConn, err := cli.connectToServer(server); err != nil {
				// TODO: log
			} else {
				cli.netConn = netConn

				break
			}
		}

		if cli.netConn == nil {
			return ErrNoAvailableServer
		}
	}

	if cli.conn == nil {
		conn, err := stomp.Connect(
			cli.netConn,
			stomp.ConnOpt.Login(cli.user, cli.pass),
		)
		if err != nil {
			return err
		}

		cli.conn = conn
	}

	cli.isConnected = true

	return nil
}

func (cli *stompClient) close() error {
	if !cli.isConnecting() {
		return nil
	}

	cli.connMutex.Lock()
	defer cli.connMutex.Unlock()

	if cli.conn != nil {
		if err := cli.conn.Disconnect(); err != nil {
			return err
		}

		cli.conn = nil
		cli.netConn = nil
	}

	if cli.netConn != nil {
		if err := cli.netConn.Close(); err != nil {
			return err
		}

		cli.netConn = nil
	}

	cli.isConnected = false

	return nil
}

func (cli *stompClient) isConnecting() bool {
	// stomp library doesn't expose closed flag, we just check connection state by isConnected flag.
	return cli.isConnected
}

func (cli *stompClient) publish(input PublishInput) error {
	destination, err := cli.getDestination(input)
	if err != nil {
		return err
	}

	opts := cli.getSendOpts(input)

	if err := cli.conn.Send(destination, input.ContentType, input.Body, opts...); err != nil {
		return err
	}

	return nil
}

func (cli *stompClient) parseServers(config Config) {

	if len(config.Servers) == 0 {
		cli.servers = make([]*stompServer, 1)
		cli.servers[0] = &stompServer{
			ssl:     false,
			address: fmt.Sprintf("127.0.0.1:%d", stompDefaultPort),
		}
		return
	}

	cli.servers = make([]*stompServer, 0, len(config.Servers))
	for _, server := range config.Servers {
		srv := new(stompServer)
		srv.ssl = server.SSL

		host := utils.Conditional(server.Host == "", "127.0.0.1", server.Host)
		port := utils.Conditional(server.Port == 0, utils.Conditional(srv.ssl, stompDefaultPortSSL, stompDefaultPort), server.Port)

		srv.address = fmt.Sprintf("%s:%d", host, port)

		cli.servers = append(cli.servers, srv)
	}
}

func (cli *stompClient) connectToServer(server *stompServer) (io.ReadWriteCloser, error) {
	if server.ssl {
		return tls.Dial("tcp", server.address, &tls.Config{})
	} else {
		return net.Dial("tcp", server.address)
	}
}

func (cli *stompClient) getDestination(input PublishInput) (string, error) {
	destination := ""

	if input.Queue != "" {
		destination = "/queue/" + input.Queue
	} else if input.Topic != "" {
		destination = "/topic/" + input.Topic
	} else {
		return "", ErrUnknownDestination
	}

	return destination, nil
}

func (cli *stompClient) getSendOpts(input PublishInput) []func(*frame.Frame) error {
	opts := make([]func(*frame.Frame) error, 0)

	if input.Expires > 0 {
		opts = append(opts, stomp.SendOpt.Header("expires", strconv.Itoa(input.Expires)))
	}
	if input.MessageId != "" {
		opts = append(opts, stomp.SendOpt.Header("message-id", input.MessageId))
	}
	if input.Persistent {
		opts = append(opts, stomp.SendOpt.Header("persistent", "true"))
	}
	if input.Priority > 0 {
		opts = append(opts, stomp.SendOpt.Header("priority", strconv.Itoa(input.Priority)))
	}

	return opts
}
