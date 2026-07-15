package message

import (
	"context"
	"encoding/json"
	"errors"
	"rakhsh/internal/common"
	"rakhsh/internal/core/client"
	"rakhsh/pkg/postgres"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/shopspring/decimal"
)

const ProcessingSubmittedMessagesInterval = 10 * time.Second
const BatchUpdatesBufferSize = 1024
const BatchUpdatesFlushInterval = 10 * time.Second
const BatchUpdatesSize = 100

type Operator interface {
	Send(message *Message) error
	Fetch(clientId int32, uids []uint64) ([]SubmittedMessage, error)
}

type MessageService struct {
	transctionManager postgres.Transactional
	clientRepository  *client.ClientRepository
	messageRepository *MessageRepository
	operatorService   Operator

	BatchUpdatesCh chan Message
}

func NewMessageService(
	transctionManager postgres.Transactional,
	clientRepository *client.ClientRepository,
	messageRepository *MessageRepository,
	operatorService Operator,
) *MessageService {
	m := &MessageService{
		transctionManager: transctionManager,
		clientRepository:  clientRepository,
		messageRepository: messageRepository,
		operatorService:   operatorService,
		BatchUpdatesCh:    make(chan Message, BatchUpdatesBufferSize),
	}

	ctx := context.Background()

	go m.batchUpdater(ctx)

	return m
}

func (m *MessageService) PostMessage(ctx context.Context, input PostMessageInput) (PostMessageOutput, error) {
	messages := make([]Message, 0, len(input.Recipients))
	messagesUids := make([]string, 0, len(input.Recipients))

	for _, recipient := range input.Recipients {
		msg, err := NewMessage(
			input.ClientId,
			recipient,
			input.Text,
			input.IsExpress,
		)

		if err != nil {
			return PostMessageOutput{}, common.InternalError("")
		}

		messages = append(messages, msg)
		messagesUids = append(messagesUids, msg.GetUidString())
	}

	err := m.transctionManager.WithinTx(ctx, input.ClientId, func(txCtx context.Context) error {
		client, err := m.clientRepository.FindClientById(txCtx, input.ClientId)
		if err != nil {
			if errors.Is(err, common.ErrNotFound) {
				return common.NotFoundError("client does not exists")
			}

			return common.InternalError("")
		}

		totalCost := common.PostMessageCost.Mul(decimal.NewFromInt(int64(len(messages))))

		if !client.IsBalanceEnough(totalCost) {
			return common.BadRequestError("insufficient balance")
		}

		if err = m.clientRepository.UpdateBalanceByAmount(txCtx, input.ClientId, totalCost.Neg()); err != nil {
			return common.InternalError("")
		}

		if err := m.messageRepository.BatchInsertMessages(txCtx, messages); err != nil {
			return common.InternalError("")
		}

		return nil
	})
	if err != nil {
		return PostMessageOutput{}, err
	}

	//TODO: need to implement outbox pattern here

	if err := m.messageRepository.PublishMessagesInQueue(ctx, messages); err != nil {
		return PostMessageOutput{
			Uids: messagesUids,
		}, nil
	}

	return PostMessageOutput{
		Uids: messagesUids,
	}, nil
}

func (m *MessageService) GetReports(ctx context.Context, clientId int32, messageUids []uint64) (GetReportsOutput, error) {
	messages, err := m.messageRepository.FindAllMessagesByUids(ctx, clientId, messageUids)
	if err != nil {
		return GetReportsOutput{}, common.InternalError("")
	}

	return GetReportsOutput{
		Messages: messages,
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

	message.SetStatus(SubmittedMessageStatus)

	err := m.operatorService.Send(message)
	if err != nil {
		reason, err := strconv.Atoi(err.Error())
		if err != nil {
			reason = int(InternalErrorMessageReason)
		}

		message.SetStatus(RejectedMessageStatus)
		message.SetReason(MessageReason(reason))

		if err := m.messageRepository.PublishMessagesInQueue(ctx, []Message{*message}); err != nil {
			delivery.Nack(false, true)
			return
		}
	}

	delivery.Ack(false)
}

func (m *MessageService) ProcessSubmittedMessage(delivery amqp.Delivery) {
	ctx := context.Background()

	//TODO: maybe its not good, wee need to do timing in another way
	time.Sleep(ProcessingSubmittedMessagesInterval)

	submittedMessage := &SubmittedMessage{}
	if err := json.Unmarshal(delivery.Body, submittedMessage); err != nil {
		delivery.Nack(false, false)
		return
	}

	message, err := m.messageRepository.FindMessageByUid(ctx, submittedMessage.ClientId, submittedMessage.Uid)
	if err != nil {
		delivery.Nack(false, true)
		return
	}

	message.SetStatus(submittedMessage.Status)
	if message.IsRejected() {
		message.SetReason(OperatorErrorMessageReason)
	}

	if publishErr := m.messageRepository.PublishMessagesInQueue(ctx, []Message{message}); publishErr != nil {
		delivery.Nack(false, true)
		return
	}

	delivery.Ack(false)
}

func (m *MessageService) ProcessDeliveredMessage(delivery amqp.Delivery) {
	ctx := context.Background()

	message := &Message{}
	if err := json.Unmarshal(delivery.Body, message); err != nil {
		delivery.Nack(false, false)
		return
	}

	if err := m.messageRepository.UpdateMessage(ctx, message); err != nil {
		delivery.Nack(false, true)
		return
	}

	m.BatchUpdatesCh <- *message

	delivery.Ack(false)
}

func (m *MessageService) ProcessRejectedMessage(delivery amqp.Delivery) {
	ctx := context.Background()

	message := &Message{}
	if err := json.Unmarshal(delivery.Body, message); err != nil {
		delivery.Nack(false, false)
		return
	}

	if err := m.messageRepository.UpdateMessage(ctx, message); err != nil {
		delivery.Nack(false, true)
		return
	}

	m.BatchUpdatesCh <- *message

	delivery.Ack(false)
}

func (m *MessageService) batchUpdater(ctx context.Context) {
	ticker := time.NewTicker(BatchUpdatesFlushInterval)
	defer ticker.Stop()

	updates := make([]Message, 0, BatchUpdatesSize)

	for {
		select {
		case <-ctx.Done():
			if len(updates) > 0 {
				m.messageRepository.BatchUpdateMessages(ctx, updates)
			}
			return

		case update := <-m.BatchUpdatesCh:
			updates = append(updates, update)

			if len(updates) >= BatchUpdatesSize {
				m.messageRepository.BatchUpdateMessages(ctx, updates)
				updates = make([]Message, 0, BatchUpdatesSize)
			}

		case <-ticker.C:
			if len(updates) > 0 {
				m.messageRepository.BatchUpdateMessages(ctx, updates)
				updates = make([]Message, 0, BatchUpdatesSize)
			}
		}
	}
}
