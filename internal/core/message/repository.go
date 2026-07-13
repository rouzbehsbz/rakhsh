package message

import (
	"context"
	"encoding/json"
	"fmt"
	postgresDb "rakhsh/db/postgres/gen"
	"rakhsh/internal/common"
	"rakhsh/pkg/postgres"
	"rakhsh/pkg/rabbitmq"
	"rakhsh/pkg/redis"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const MessageCacheTTL = 10 * time.Minute

type MessageRepository struct {
	db    *postgres.Postgres
	queue *rabbitmq.Rabbitmq
	cache *redis.Redis
}

func NewMessageRepository(db *postgres.Postgres, queue *rabbitmq.Rabbitmq, cache *redis.Redis) *MessageRepository {
	return &MessageRepository{
		db:    db,
		queue: queue,
		cache: cache,
	}
}

func (m *MessageRepository) InsertMessage(ctx context.Context, message *Message) error {
	shard := m.db.GetShard(message.ClientId)
	q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

	err := q.InsertMessage(ctx, postgresDb.InsertMessageParams(MapMessageToPgMessage(message)))
	if err != nil {
		return common.ErrInternalDatabase
	}

	_ = m.cache.SetJSON(
		ctx,
		messageKey(message.ClientId, message.Uid),
		message,
		MessageCacheTTL,
	)

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

	_ = m.cache.Delete(
		ctx,
		messageKey(message.ClientId, message.Uid),
	)

	return nil
}

func (m *MessageRepository) FindMessageByUid(ctx context.Context, clientId int32, uid uint64) (Message, error) {
	key := messageKey(clientId, uid)

	var cached Message

	err := m.cache.GetJson(ctx, key, &cached)
	if err == nil {
		return cached, nil
	}

	shard := m.db.GetShard(clientId)
	q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

	pgMessage, err := q.FindMessageByUid(
		ctx,
		postgresDb.FindMessageByUidParams{
			ClientID: clientId,
			Uid:      int64(uid),
		},
	)

	if err != nil {
		return Message{}, common.ErrInternalDatabase
	}

	message := MapPgMessageToMessage(&pgMessage)

	_ = m.cache.SetJSON(
		ctx,
		key,
		message,
		MessageCacheTTL,
	)

	return message, nil
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

func messageKey(clientId int32, uid uint64) string {
	return fmt.Sprintf("client:%d:message:%d", clientId, uid)
}
