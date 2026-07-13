package message

import (
	postgresDb "rakhsh/db/postgres/gen"
	"rakhsh/internal/common"
	"rakhsh/pkg/postgres"
	"strconv"
	"time"

	"github.com/godruoyi/go-snowflake"
)

type MessageStatus int16

const (
	PendingMessageStatus MessageStatus = iota
	SubmittedMessageStatus
	DeliveredMessageStatus
	RejectedMessageStatus
)

type MessageReason int16

const (
	NullMessageReason MessageReason = iota
	InternalErrorMessageReason
	OperatorErrorMessageReason
)

type Message struct {
	Uid       uint64        `json:"uid"`
	ClientId  int32         `json:"clientId"`
	Recipient string        `json:"recipient"`
	Text      string        `json:"text"`
	IsExpress bool          `json:"isExpress"`
	Status    MessageStatus `json:"status"`
	Reason    MessageReason `json:"reason"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

func NewMessage(clientId int32, recipient, text string, isExpress bool) (Message, error) {
	uid, err := snowflake.NextID()
	if err != nil {
		return Message{}, err
	}

	return Message{
		Uid:       uid,
		ClientId:  clientId,
		Recipient: recipient,
		Text:      text,
		IsExpress: isExpress,
		Status:    PendingMessageStatus,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *Message) GetUidString() string {
	str := strconv.FormatUint(m.Uid, 10)
	return str
}

func (m *Message) IsPending() bool {
	if m.Status == PendingMessageStatus {
		return true
	}

	return false
}

func (m *Message) IsSubmitted() bool {
	if m.Status == SubmittedMessageStatus {
		return true
	}

	return false
}

func (m *Message) IsRejected() bool {
	if m.Status == RejectedMessageStatus {
		return true
	}

	return false
}

func (m *Message) IsDelivered() bool {
	if m.Status == DeliveredMessageStatus {
		return true
	}

	return false
}

func (m *Message) SetStatus(status MessageStatus) {
	m.Status = status
}

func (m *Message) SetReason(reason MessageReason) {
	m.Reason = reason
}

func (m *Message) GetQueueName() string {
	switch m.Status {
	case PendingMessageStatus:
		return common.PendingMessagesQueueName
	case SubmittedMessageStatus:
		return common.SubmittedMessagesQueueName
	case DeliveredMessageStatus:
		return common.DeliveredMessageQueueName
	case RejectedMessageStatus:
		return common.RejectedMessagesQueueName
	default:
		return ""
	}
}

func MapMessageToPgMessage(message *Message) postgresDb.Message {
	return postgresDb.Message{
		Uid:       int64(message.Uid),
		CreatedAt: postgres.TimeToPgTimestampz(message.CreatedAt),
		UpdatedAt: postgres.TimeToPgTimestampz(message.UpdatedAt),
		ClientID:  message.ClientId,
		Status:    int16(message.Status),
		Reason:    postgres.IntToPgInt2(int16(message.Reason), false),
		IsExpress: message.IsExpress,
		Recipient: message.Recipient,
		Text:      message.Text,
	}
}

func MapPgMessageToMessage(pgMessage *postgresDb.Message) Message {
	return Message{
		Uid:       uint64(pgMessage.Uid),
		ClientId:  pgMessage.ClientID,
		Recipient: pgMessage.Recipient,
		Text:      pgMessage.Text,
		IsExpress: pgMessage.IsExpress,
		Status:    MessageStatus(pgMessage.Status),
		Reason:    MessageReason(pgMessage.Reason.Int16),
		CreatedAt: pgMessage.CreatedAt.Time,
		UpdatedAt: pgMessage.UpdatedAt.Time,
	}
}

type SubmittedMessage struct {
	Uid      uint64        `json:"uid"`
	ClientId int32         `json:"clientId"`
	Status   MessageStatus `json:"status"`
}
