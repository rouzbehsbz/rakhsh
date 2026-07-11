package main

import (
	"log"
	"rakhsh/internal/api"
	"rakhsh/internal/common"
	"rakhsh/internal/core/client"
	"rakhsh/pkg/postgres"
)

func main() {
	config, err := common.NewConfig(true)
	if err != nil {
		panic(err)
	}

	postgres, err := postgres.NewPostgresService(
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresDatabaseName,
		config.PostgresMaxConnections,
	)
	if err != nil {
		panic(err)
	}

	clientRepository := client.NewClientRepository(postgres.Q)
	clientService := client.NewClientService(clientRepository)
	clientHandler := client.NewClientHandler(clientService)

	server := api.NewServer(config.Host, config.Port, api.ServerOpts{
		ClientHandler: clientHandler,
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
