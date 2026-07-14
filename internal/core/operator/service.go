package operator

import (
	"fmt"
	"math/rand"
	"rakhsh/internal/core/message"
	"sync/atomic"
	"time"
)

type OperatorService struct {
	operators []message.Operator
	index     atomic.Int32
}

func NewOperatorService() *OperatorService {
	return &OperatorService{
		operators: []message.Operator{},
		index:     atomic.Int32{},
	}
}

func (o *OperatorService) RegisterOperator(operator message.Operator) {
	o.operators = append(o.operators, operator)
}

func (o *OperatorService) nextIndex() int32 {
	index := o.index.Add(1)
	return index % int32(len(o.operators))
}

func (o *OperatorService) nextOperator() message.Operator {
	nextIndex := o.nextIndex()
	return o.operators[nextIndex]
}

func (o *OperatorService) Send(message *message.Message) error {
	operator := o.nextOperator()

	return operator.Send(message)
}

func (o *OperatorService) Fetch(clientId int32, uids []uint64) ([]message.SubmittedMessage, error) {
	return nil, fmt.Errorf("service itself doesn't fetch any data")
}

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
