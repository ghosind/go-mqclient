package mqclient

import "errors"

var ErrUnsupportedProtocol = errors.New("unsupported protocol")
var ErrNoAvailableServer = errors.New("no available server")
