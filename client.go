package mqclient

import (
	"strings"
)

type client interface {
	connect() error
	close() error
	isConnecting() bool
}

func newClientByProtocol(protocol string, config Config) (client, error) {
	switch strings.TrimSpace(strings.ToLower(config.Protocol)) {
	case ProtocolAMQP091:
		return newAmqp091Client(config)
	default:
		return nil, ErrUnsupportedProtocol
	}
}
