package mqclient

const (
	// Protocol AMQP 0.9.1
	ProtocolAMQP091 = "amqp091"
)

type ServerConfig struct {
	SSL  bool
	Host string
	Port int
}

type Config struct {
	Protocol      string
	Servers       []ServerConfig
	User          string
	Pass          string
	VHost         string
	AutoConnect   bool
	AutoReconnect bool
}
