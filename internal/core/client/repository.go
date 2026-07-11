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
	q *postgresDb.Queries
}

func NewClientRepository(q *postgresDb.Queries) *ClientRepository {
	return &ClientRepository{
		q: q,
	}
}

func (c *ClientRepository) FindClientById(ctx context.Context, id int32) (Client, error) {
	pgClient, err := c.q.FindClientById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Client{}, common.ErrNotFound
		}

		return Client{}, common.ErrInternalDatabase
	}

	return MapPgClientToClient(&pgClient)
}

func (c *ClientRepository) UpdateBalanceByAmount(ctx context.Context, id int32, amount decimal.Decimal) error {
	balance, err := postgres.MapDecimalToPgNumeric(amount)
	if err != nil {
		return err
	}

	err = c.q.UpdateBalanceByAmount(ctx, postgresDb.UpdateBalanceByAmountParams{
		ID:      id,
		Balance: balance,
	})
	if err != nil {
		return common.ErrInternalDatabase
	}

	return nil
}
