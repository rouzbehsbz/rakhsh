package client

import (
	"context"
	"errors"
	postgresDb "rakhsh/db/postgres/gen"
	"rakhsh/internal/common"
	"rakhsh/pkg/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type ClientRepository struct {
	db *postgres.Postgres
}

func NewClientRepository(db *postgres.Postgres) *ClientRepository {
	return &ClientRepository{
		db: db,
	}
}

func (c *ClientRepository) FindClientById(ctx context.Context, id int32) (Client, error) {
	shard := c.db.GetShard(id)
	q := postgres.ExtractTxQuery(shard.MasterQ, ctx)

	pgClient, err := q.FindClientById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Client{}, common.ErrNotFound
		}

		return Client{}, common.ErrInternalDatabase
	}

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

	return nil
}
