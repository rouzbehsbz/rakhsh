package message

import (
	"context"
	"encoding/json"
	postgresDb "rakhsh/db/postgres/gen"
	"rakhsh/internal/common"
	"rakhsh/pkg/postgres"
	"rakhsh/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageRepository struct {
	db    *postgres.Postgres
	queue *rabbitmq.Rabbitmq
}

func NewMessageRepository(db *postgres.Postgres, queue *rabbitmq.Rabbitmq) *MessageRepository {
	return &MessageRepository{
		db:    db,
		queue: queue,
	}
}

func (m *MessageRepository) InsertMessage(ctx context.Context, message *Message) error {
	shard := m.db.GetShard(message.ClientId)
	q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

	err := q.InsertMessage(ctx, postgresDb.InsertMessageParams(MapMessageToPgMessage(message)))
	if err != nil {
		return common.ErrInternalDatabase
	}

	return nil
}

func (m *MessageRepository) UpdateMessage(ctx context.Context, message *Message) error {
	shard := m.db.GetShard(message.ClientId)
	q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

	pgMessage := MapMessageToPgMessage(message)

	err := q.UpdateMessage(ctx, postgresDb.UpdateMessageParams{
		Status: pgMessage.Status,
		Reason: pgMessage.Reason,
		Uid:    pgMessage.Uid,
	})
	if err != nil {
		return common.ErrInternalDatabase
	}

	return nil
}

func (m *MessageRepository) PublishMessageInQueue(ctx context.Context, message *Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	priority := uint8(1)
	if message.IsExpress {
		priority = 5
	}

	return m.queue.Publish(ctx, message.GetQueueName(), amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
		Priority:    priority,
	})
}
