package message

import (
	"context"
	postgresDb "rakhsh/db/postgres/gen"
	"rakhsh/internal/common"
	"rakhsh/pkg/postgres"
)

type MessageRepository struct {
	db *postgres.Postgres
}

func NewMessageRepository(db *postgres.Postgres) *MessageRepository {
	return &MessageRepository{
		db: db,
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
