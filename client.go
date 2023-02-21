package mqclient

import (
	"strings"
)

type PublishInput struct {
	Body        []byte
	ContentType string
	Queue       string
	Topic       string
}

type client interface {
	connect() error
	close() error
	isConnecting() bool
	publish(PublishInput) error
}

func newClientByProtocol(protocol string, config Config) (client, error) {
	switch strings.TrimSpace(strings.ToLower(config.Protocol)) {
	case ProtocolAMQP091:
		return newAmqp091Client(config)
	case ProtocolSTOMP:
		return newStompClient(config)
	default:
		return nil, ErrUnsupportedProtocol
	}
}
