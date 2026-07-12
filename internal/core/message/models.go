package message

import (
	postgresDb "rakhsh/db/postgres/gen"
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
	InternalErrorMessageReason MessageReason = iota
	OperatorErrorMessageReason
)

type Message struct {
	Uid       uint64
	ClientId  int32
	Recipient string
	Text      string
	IsExpress bool
	Status    MessageStatus
	Reason    MessageReason
	CreatedAt time.Time
	UpdatedAt time.Time
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

func MapMessageToPgMessage(message *Message) postgresDb.Message {
	return postgresDb.Message{
		Uid:       int64(message.Uid),
		CreatedAt: postgres.TimeToPgTimestampz(message.CreatedAt),
		UpdatedAt: postgres.TimeToPgTimestampz(message.UpdatedAt),
		ClientID:  message.ClientId,
		Status:    int16(message.Status),
		Reason:    postgres.IntToPgInt2(0, false),
		IsExpress: message.IsExpress,
		Recipient: message.Recipient,
		Text:      message.Text,
	}
}
