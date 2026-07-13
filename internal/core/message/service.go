package message

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"rakhsh/internal/common"
	"rakhsh/internal/core/client"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/shopspring/decimal"
)

type Operator interface {
	Send(message *Message) error
}

type MessageService struct {
	transctionManager common.Transactional
	clientRepository  *client.ClientRepository
	messageRepository *MessageRepository
	operatorService   Operator
}

func NewMessageService(
	transctionManager common.Transactional,
	clientRepository *client.ClientRepository,
	messageRepository *MessageRepository,
	operatorService Operator,
) *MessageService {
	return &MessageService{
		transctionManager: transctionManager,
		clientRepository:  clientRepository,
		messageRepository: messageRepository,
		operatorService:   operatorService,
	}
}

func (m *MessageService) PostMessage(ctx context.Context, input PostMessageInput) (PostMessageOutput, error) {
	var message Message

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

		message, err = NewMessage(input.ClientId, input.Recipient, input.Text, input.IsExpress)
		if err != nil {
			return common.InternalError("")
		}

		if err := m.messageRepository.InsertMessage(txCtx, &message); err != nil {
			return common.InternalError("")
		}

		return nil
	})
	if err != nil {
		return PostMessageOutput{}, err
	}

	if err := m.messageRepository.PublishMessageInQueue(ctx, &message); err != nil {
		return PostMessageOutput{
			Uid: message.GetUidString(),
		}, nil
	}

	return PostMessageOutput{
		Uid: message.GetUidString(),
	}, nil
}

func (m *MessageService) ProcessPendingMessage(delivery amqp.Delivery) {
	ctx := context.Background()

	message := &Message{}
	if err := json.Unmarshal(delivery.Body, message); err != nil {
		delivery.Nack(false, false)
		return
	}

	if !message.IsPending() {
		delivery.Ack(false)
		return
	}

	err := m.transctionManager.WithinTx(ctx, message.ClientId, func(txCtx context.Context) error {
		message.SetStatus(SubmittedMessageStatus)

		if err := m.messageRepository.UpdateMessage(txCtx, message); err != nil {
			return common.InternalError(fmt.Sprintf("%d", InternalErrorMessageReason))
		}

		if err := m.operatorService.Send(message); err != nil {
			return common.InternalError(fmt.Sprintf("%d", OperatorErrorMessageReason))
		}

		return nil
	})
	if err != nil {
		reason, err := strconv.Atoi(err.Error())
		if err != nil {
			reason = int(InternalErrorMessageReason)
		}

		message.SetStatus(RejectedMessageStatus)
		message.SetReason(MessageReason(reason))

		if updateErr := m.messageRepository.UpdateMessage(ctx, message); updateErr != nil {
			delivery.Nack(false, true)
			return
		}

		if publishErr := m.messageRepository.PublishMessageInQueue(ctx, message); publishErr != nil {
			delivery.Nack(false, true)
			return
		}
	}

	delivery.Ack(false)
}

func (m *MessageService) ProcessSubmittedMessage(delivery amqp.Delivery) {
	ctx := context.Background()

	dMessage := &Message{}
	if err := json.Unmarshal(delivery.Body, dMessage); err != nil {
		delivery.Nack(false, false)
		return
	}

	if !dMessage.IsSubmitted() {
		delivery.Ack(false)
		return
	}

	lMessage, err := m.messageRepository.FindMessageByUid(ctx, dMessage.ClientId, dMessage.Uid)
	if err != nil {
		delivery.Nack(false, true)
		return
	}

	if !lMessage.IsDelivered() {
		delivery.Nack(false, true)
	}

	if publishErr := m.messageRepository.PublishMessageInQueue(ctx, &lMessage); publishErr != nil {
		delivery.Nack(false, true)
		return
	}

	delivery.Ack(false)
}

func (m *MessageService) ProcessDeliveredMessage(delivery amqp.Delivery) {
	delivery.Ack(false)
}

func (m *MessageService) ProcessRejectedMessage(delivery amqp.Delivery) {
	delivery.Ack(false)
}

func (m *MessageService) GetMessage(ctx context.Context, clientId int, messageId string) (GetMessageOutput, error) {
	return GetMessageOutput{}, nil
}
