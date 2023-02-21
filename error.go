package mqclient

import "errors"

var ErrNoAvailableServer = errors.New("no available server")
var ErrUnknownDestination = errors.New("unknown destination")
var ErrUnsupportedProtocol = errors.New("unsupported protocol")
