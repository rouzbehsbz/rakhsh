package message

import (
	"context"
	"errors"
	"rakhsh/internal/common"
	"rakhsh/internal/core/client"

	"github.com/shopspring/decimal"
)

type MessageService struct {
	transctionManager common.Transactional
	clientRepository  *client.ClientRepository
	messageRepository *MessageRepository
}

func NewMessageService(transctionManager common.Transactional, clientRepository *client.ClientRepository, messageRepository *MessageRepository) *MessageService {
	return &MessageService{
		transctionManager: transctionManager,
		clientRepository:  clientRepository,
		messageRepository: messageRepository,
	}
}

func (m *MessageService) PostMessage(ctx context.Context, input PostMessageInput) (PostMessageOutput, error) {
	var messageUid string

	err := m.transctionManager.WithinTx(ctx, input.ClientId, func(txCtx context.Context) error {
		client, err := m.clientRepository.FindClientById(txCtx, input.ClientId)
		if err != nil {
			if errors.Is(err, common.ErrNotFound) {
				return common.NotFoundError("client does not exists")
			}

			return common.InternalError("")
		}

		cost := common.PostMessageCost

		if ok := client.IsBalanceEnough(cost); !ok {
			return common.BadRequestError("insufficient balance")
		}

		if err = m.clientRepository.UpdateBalanceByAmount(txCtx, input.ClientId, cost.Mul(decimal.NewFromInt(-1))); err != nil {
			return common.InternalError("")
		}

		message, err := NewMessage(input.ClientId, input.Recipient, input.Text, input.IsExpress)
		if err != nil {
			return common.InternalError("")
		}

		if err := m.messageRepository.InsertMessage(txCtx, &message); err != nil {
			return common.InternalError(err.Error())
		}

		messageUid = message.GetUidString()

		return nil
	})
	if err != nil {
		return PostMessageOutput{}, err
	}

	return PostMessageOutput{
		Uid: messageUid,
	}, nil
}

func (m *MessageService) GetMessage(ctx context.Context, clientId int, messageId string) (GetMessageOutput, error) {
	return GetMessageOutput{}, nil
}
