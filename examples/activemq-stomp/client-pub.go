package main

import (
	"log"

	"github.com/ghosind/go-mqclient"
)

func main() {
	cli, err := mqclient.New(mqclient.Config{
		Protocol: mqclient.ProtocolSTOMP,
		Servers: []mqclient.ServerConfig{
			{
				Host: "127.0.0.1",
				Port: 61613,
			},
		},
		User:        "guest",
		Pass:        "guest",
		AutoConnect: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to ActiveMQ: %v", err)
	}

	// TODO: publish a message

	if err := cli.Close(); err != nil {
		log.Fatalf("Failed to close connection: %v", err)
	}
}
