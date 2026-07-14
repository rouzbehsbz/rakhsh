package operator

import (
	"fmt"
	"rakhsh/internal/core/message"
	"sync/atomic"
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
