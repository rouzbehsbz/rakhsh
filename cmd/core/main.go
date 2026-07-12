package main

import (
	"log"
	"rakhsh/internal/api"
	"rakhsh/internal/common"
	"rakhsh/internal/core/client"
	"rakhsh/internal/core/message"
	"rakhsh/pkg/postgres"

	"github.com/godruoyi/go-snowflake"
)

func main() {
	config, err := common.NewConfig(true)
	if err != nil {
		panic(err)
	}

	snowflake.SetMachineID(config.MachineId)

	celebritiesShard := map[int32]int{}

	postgres, err := postgres.NewPostgresService([]string{
		config.PostgresShard1,
		config.PostgresShard2,
		config.PostgresShard3,
	}, celebritiesShard, config.PostgresMaxConnections)
	if err != nil {
		panic(err)
	}

	clientRepository := client.NewClientRepository(postgres)
	messageRepository := message.NewMessageRepository(postgres)

	clientService := client.NewClientService(clientRepository)
	messageService := message.NewMessageService(postgres, clientRepository, messageRepository)

	clientHandler := client.NewClientHandler(clientService)
	messageHandler := message.NewMessageHandler(messageService)

	server := api.NewServer(config.Host, config.Port, api.RootHandlers{
		ClientHandler:  clientHandler,
		MessageHandler: messageHandler,
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
