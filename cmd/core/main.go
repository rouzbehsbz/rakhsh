package main

import (
	"log"
	"rakhsh/internal/api"
	"rakhsh/internal/common"
	"rakhsh/internal/core/client"
	"rakhsh/internal/core/message"
	"rakhsh/internal/core/operator"
	"rakhsh/pkg/postgres"
	"rakhsh/pkg/rabbitmq"
	"rakhsh/pkg/redis"

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
		config.PostgresShard1Url,
		config.PostgresShard2Url,
		config.PostgresShard3Url,
	}, celebritiesShard, config.PostgresMaxConnections)
	if err != nil {
		panic(err)
	}

	rabbitmq, err := rabbitmq.NewRabbitmq(config.RabbitmqUrl)
	if err != nil {
		panic(err)
	}

	redis, err := redis.NewRedis(config.RedisUrl, config.RedisPassword, config.RedisMaxConnections)
	if err != nil {
		panic(err)
	}

	clientRepository := client.NewClientRepository(postgres, redis)
	messageRepository := message.NewMessageRepository(postgres, rabbitmq, redis)

	operatorService := operator.NewOperatorService()
	operatorService.RegisterOperator(operator.NewDummyOperator())
	operatorService.RegisterOperator(operator.NewDummyOperator())
	operatorService.RegisterOperator(operator.NewDummyOperator())

	clientService := client.NewClientService(clientRepository)
	messageService := message.NewMessageService(postgres, clientRepository, messageRepository, operatorService)

	clientHandler := client.NewClientHandler(clientService)
	messageHandler := message.NewMessageHandler(messageService)

	if err := rabbitmq.AddQueue(common.PendingMessagesQueueName, messageService.ProcessPendingMessage); err != nil {
		panic(err)
	}
	if err := rabbitmq.AddQueue(common.SubmittedMessagesQueueName, messageService.ProcessSubmittedMessage); err != nil {
		panic(err)
	}
	if err := rabbitmq.AddQueue(common.DeliveredMessageQueueName, messageService.ProcessDeliveredMessage); err != nil {
		panic(err)
	}
	if err := rabbitmq.AddQueue(common.RejectedMessagesQueueName, messageService.ProcessRejectedMessage); err != nil {
		panic(err)
	}

	if err := rabbitmq.StartQueueConsumers(common.PendingMessagesQueueName, config.RabbitmqMaxWorkers); err != nil {
		panic(err)
	}

	server := api.NewServer(config.Host, config.Port, api.RootHandlers{
		ClientHandler:  clientHandler,
		MessageHandler: messageHandler,
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
