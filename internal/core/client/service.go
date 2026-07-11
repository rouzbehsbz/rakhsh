package client

import (
	"context"
	"errors"
	"rakhsh/internal/common"

	"github.com/shopspring/decimal"
)

type ClientService struct {
	repository *ClientRepository
}

func NewClientService(repository *ClientRepository) *ClientService {
	return &ClientService{
		repository: repository,
	}
}

func (c *ClientService) GetClientInfo(ctx context.Context, id int32) (GetClientInfoOutput, error) {
	client, err := c.repository.FindClientById(ctx, id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return GetClientInfoOutput{}, common.NotFoundError("Client does not exists")
		}

		return GetClientInfoOutput{}, common.InternalError("")
	}

	return GetClientInfoOutput{
		Name:    client.Name,
		Balance: client.Balance,
	}, nil
}

func (c *ClientService) ChargeBalance(ctx context.Context, id int32, amount decimal.Decimal) error {
	_, err := c.repository.FindClientById(ctx, id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return common.NotFoundError("Client does not exists")
		}

		return common.InternalError("")
	}

	err = c.repository.UpdateBalanceByAmount(ctx, id, amount)
	if err != nil {
		return common.InternalError("")
	}

	return nil
}
