package main

import (
	"log"

	"github.com/ghosind/go-mqclient"
)

func main() {
	cli, err := mqclient.New(mqclient.Config{
		Protocol: mqclient.ProtocolAMQP091,
		Servers: []mqclient.ServerConfig{
			{
				Host: "127.0.0.1",
				Port: 5672,
			},
		},
		User:        "guest",
		Pass:        "guest",
		VHost:       "/",
		AutoConnect: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	// TODO: publish a message

	if err := cli.Close(); err != nil {
		log.Fatalf("Failed to close connection: %v", err)
	}
}
