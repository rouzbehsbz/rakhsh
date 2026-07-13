package client

import (
	"context"
	"errors"
	"fmt"
	postgresDb "rakhsh/db/postgres/gen"
	"rakhsh/internal/common"
	"rakhsh/pkg/postgres"
	"rakhsh/pkg/redis"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

const ClientCacheTTL = 10 * time.Minute

type ClientRepository struct {
	db    *postgres.Postgres
	cache *redis.Redis
}

func NewClientRepository(db *postgres.Postgres, cache *redis.Redis) *ClientRepository {
	return &ClientRepository{
		db:    db,
		cache: cache,
	}
}

func (c *ClientRepository) FindClientById(ctx context.Context, id int32) (Client, error) {
	key := clientKey(id)

	var client Client
	if err := c.cache.GetJson(ctx, key, &client); err != common.ErrNotFound {
		return client, nil
	}

	shard := c.db.GetShard(id)
	q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

	pgClient, err := q.FindClientById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Client{}, common.ErrNotFound
		}

		return Client{}, common.ErrInternalDatabase
	}

	client, err = MapPgClientToClient(&pgClient)
	if err != nil {
		return Client{}, err
	}

	_ = c.cache.SetJSON(ctx, key, client, ClientCacheTTL)

	return MapPgClientToClient(&pgClient)
}

func (c *ClientRepository) UpdateBalanceByAmount(ctx context.Context, id int32, amount decimal.Decimal) error {
	shard := c.db.GetShard(id)
	q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

	balance, err := postgres.MapDecimalToPgNumeric(amount)
	if err != nil {
		return err
	}

	err = q.UpdateBalanceByAmount(ctx, postgresDb.UpdateBalanceByAmountParams{
		ID:      id,
		Balance: balance,
	})
	if err != nil {
		return common.ErrInternalDatabase
	}

	_ = c.cache.Delete(ctx, clientKey(id))

	return nil
}

func clientKey(id int32) string {
	return fmt.Sprintf("client:%d", id)
}
