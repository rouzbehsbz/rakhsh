package operator

import (
	"fmt"
	"math/rand"
	"rakhsh/internal/core/message"
	"time"
)

type DummyOperator struct{}

func NewDummyOperator() *DummyOperator {
	return &DummyOperator{}
}

func (d *DummyOperator) Send(message *message.Message) error {
	n := rand.Intn(3)
	t := 30 + rand.Intn(70)

	var err error
	if n == 0 {
		err = fmt.Errorf("can't reach the operator")
	}

	time.Sleep(time.Duration(t) * time.Millisecond)

	return err
}

func (d *DummyOperator) Fetch(clientId int32, uids []uint64) ([]message.SubmittedMessage, error) {
	n := rand.Intn(5)
	if n == 0 {
		return nil, fmt.Errorf("faild to fetch data")
	}

	count := len(uids)
	res := make([]message.SubmittedMessage, 0, count)

	for _, uid := range uids {
		var status message.MessageStatus

		s := rand.Intn(5)
		if s == 0 {
			status = message.RejectedMessageStatus
		}

		res = append(res, message.SubmittedMessage{
			Uid:      uid,
			ClientId: clientId,
			Status:   status,
		})
	}

	return res, nil
}
