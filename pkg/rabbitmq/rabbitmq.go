package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const PublishRequestBufferedChannelSize = 1024

type QueueHandler func(amqp.Delivery)
type QueueOptions struct {
	Handler     QueueHandler
	MaxPriority int
}

type Queue struct {
	conn *amqp.Connection
	q    amqp.Queue

	publishCh *amqp.Channel
	consumeCh []*amqp.Channel

	Handler QueueHandler
}

type Rabbitmq struct {
	url    string
	queues map[string]*Queue
}

func NewRabbitmq(url string) (*Rabbitmq, error) {
	r := &Rabbitmq{
		url:    url,
		queues: make(map[string]*Queue),
	}

	return r, nil
}

func (r *Rabbitmq) AddQueue(name string, opt QueueOptions) error {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to connect to rabbitmq server")
	}

	publishCh, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("can't create publish channel for %s", name)
	}

	if err := publishCh.Confirm(false); err != nil {
		return fmt.Errorf("can't set confitm mode for consume channel for %s", name)
	}

	q, err := publishCh.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-max-priority": int32(opt.MaxPriority),
		},
	)
	if err != nil {
		return fmt.Errorf("can't declare queue for %s", name)
	}

	r.queues[name] = &Queue{
		conn:      conn,
		q:         q,
		publishCh: publishCh,
		Handler:   opt.Handler,
	}

	return nil
}

func (r *Rabbitmq) StartQueueConsumers(name string, count int) error {
	queue, ok := r.queues[name]
	if !ok {
		return fmt.Errorf("can't find queue name %s to add consuemers", name)
	}

	for range count {
		consumeCh, err := queue.conn.Channel()
		if err != nil {
			return fmt.Errorf("can't create consume channel for %s", name)
		}

		err = consumeCh.Qos(
			10,
			0,
			false,
		)
		if err != nil {
			return fmt.Errorf("can't initiate Qos for %s queue", name)
		}

		deliveries, err := consumeCh.Consume(
			name,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("can't create deliveries channel for %s queue", name)
		}

		go func() {
			//TODO: need to handle close signals
			for msg := range deliveries {
				queue.Handler(msg)
			}
		}()
	}

	return nil
}

func (r *Rabbitmq) Publish(ctx context.Context, queueName string, publishing amqp.Publishing) error {
	queue, ok := r.queues[queueName]
	if !ok {
		return fmt.Errorf("can't find queue")
	}

	return queue.publishCh.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		publishing,
	)
}
