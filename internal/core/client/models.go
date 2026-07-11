package client

import (
	"fmt"
	postgresDb "rakhsh/db/postgres/gen"
	"rakhsh/pkg/postgres"

	"github.com/shopspring/decimal"
)

type Client struct {
	Id      int32
	Name    string
	Balance decimal.Decimal
}

func MapPgClientToClient(pgClient *postgresDb.Client) (Client, error) {
	if pgClient == nil {
		return Client{}, fmt.Errorf("value is nil")
	}

	balance, err := postgres.MapPgNumericToDecimal(pgClient.Balance)
	if err != nil {
		return Client{}, err
	}

	return Client{
		Id:      pgClient.ID,
		Name:    pgClient.Name,
		Balance: balance,
	}, nil
}
