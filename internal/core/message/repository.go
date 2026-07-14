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

func (m *MessageRepository) BatchUpdateMessages(ctx context.Context, updates []Message) error {
	shards := make(map[int32][]Message)
	for _, msg := range updates {
		shards[msg.ClientId] = append(shards[msg.ClientId], msg)
	}

	for clientId, shardMessages := range shards {
		shard := m.db.GetShard(clientId)
		q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

		uids := make([]int64, len(shardMessages))
		statuses := make([]int16, len(shardMessages))
		reasons := make([]int16, len(shardMessages))
		cacheKeys := make([]string, len(shardMessages))

		for i, msg := range shardMessages {
			uids[i] = int64(msg.Uid)
			statuses[i] = int16(msg.Status)
			reasons[i] = int16(msg.Reason)

			cacheKeys[i] = messageKey(msg.ClientId, msg.Uid)
		}

		err := q.BatchUpdateMessages(ctx, postgresDb.BatchUpdateMessagesParams{
			Column1: uids,
			Column2: statuses,
			Column3: reasons,
		})
		if err != nil {
			return common.ErrInternalDatabase
		}

		_ = m.cache.DeleteMany(ctx, cacheKeys...)
	}

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

func (m *MessageRepository) FindAllMessagesByUids(ctx context.Context, clientId int32, uids []uint64) ([]Message, error) {
	messages := make([]Message, 0, len(uids))
	missing := make([]uint64, 0)

	for _, uid := range uids {
		key := messageKey(clientId, uid)

		var cached Message
		if err := m.cache.GetJson(ctx, key, &cached); err == nil {
			messages = append(messages, cached)
			continue
		}

		missing = append(missing, uid)
	}

	if len(missing) == 0 {
		return messages, nil
	}

	pgUids := make([]int64, len(missing))
	for i, uid := range missing {
		pgUids[i] = int64(uid)
	}

	shard := m.db.GetShard(clientId)
	q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

	pgMessages, err := q.FindAllMessagesByUids(ctx, postgresDb.FindAllMessagesByUidsParams{
		ClientID: clientId,
		Column2:  pgUids,
	})
	if err != nil {
		return nil, common.ErrInternalDatabase
	}

	for _, pgMessage := range pgMessages {
		message := MapPgMessageToMessage(&pgMessage)

		_ = m.cache.SetJSON(
			ctx,
			messageKey(clientId, message.Uid),
			message,
			MessageCacheTTL,
		)

		messages = append(messages, message)
	}

	return messages, nil
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

func (m *MessageRepository) PublishSubmittedMessageInQueue(ctx context.Context, submittedMessage *SubmittedMessage) error {
	body, err := json.Marshal(submittedMessage)
	if err != nil {
		return err
	}

	return m.queue.Publish(ctx, common.SubmittedMessagesQueueName, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func messageKey(clientId int32, uid uint64) string {
	return fmt.Sprintf("client:%d:message:%d", clientId, uid)
}
