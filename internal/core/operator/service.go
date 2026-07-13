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

type DummyOperator struct{}

func NewDummyOperator() *DummyOperator {
	return &DummyOperator{}
}

func (d *DummyOperator) Send(message *message.Message) error {
	n := rand.Intn(2)
	t := 30 + rand.Intn(70)

	var err error
	if n == 0 {
		err = fmt.Errorf("can't reach the operator")
	}

	time.Sleep(time.Duration(t) * time.Millisecond)

	return err
}
