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
		log.Fatalf("failed to load config: %v", err)
	}

	snowflake.SetMachineID(config.MachineId)

	celebritiesShard := map[int32]int{}

	postgres, err := postgres.NewPostgresService([]string{
		config.PostgresShard1Url,
		config.PostgresShard2Url,
		config.PostgresShard3Url,
	}, celebritiesShard, config.PostgresMaxConnections)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}

	rabbit, err := rabbitmq.NewRabbitmq(config.RabbitmqUrl)
	if err != nil {
		log.Fatalf("failed to init rabbitmq: %v", err)
	}

	redis, err := redis.NewRedis(config.RedisUrl, config.RedisPassword, config.RedisMaxConnections)
	if err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}

	clientRepository := client.NewClientRepository(postgres, redis)
	messageRepository := message.NewMessageRepository(postgres, rabbit, redis)

	operatorService := operator.NewOperatorService()
	for i := 0; i < 3; i++ {
		operatorService.RegisterOperator(operator.NewDummyOperator())
	}

	clientService := client.NewClientService(clientRepository)
	messageService := message.NewMessageService(postgres, clientRepository, messageRepository, operatorService)

	queues := []struct {
		name    string
		handler rabbitmq.QueueHandler
	}{
		{common.PendingMessagesQueueName, messageService.ProcessPendingMessage},
		{common.SubmittedMessagesQueueName, messageService.ProcessSubmittedMessage},
		{common.DeliveredMessageQueueName, messageService.ProcessDeliveredMessage},
		{common.RejectedMessagesQueueName, messageService.ProcessRejectedMessage},
	}

	for _, q := range queues {
		if err := rabbit.AddQueue(q.name, q.handler); err != nil {
			log.Fatalf("failed to add queue %s: %v", q.name, err)
		}
		if err := rabbit.StartQueueConsumers(q.name, config.RabbitmqMaxWorkers); err != nil {
			log.Fatalf("failed to start consumers for %s: %v", q.name, err)
		}
	}

	server := api.NewServer(config.Host, config.Port, api.RootHandlers{
		ClientHandler:  client.NewClientHandler(clientService),
		MessageHandler: message.NewMessageHandler(messageService),
	})

	log.Printf("Starting server on %s:%d ...", config.Host, config.Port)
	if err := server.Run(); err != nil {
		log.Fatalf("server terminated unexpectedly: %v", err)
	}
}
