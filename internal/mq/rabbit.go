package mq

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/streadway/amqp"
)

// Queue -
type Queue struct {
	Name       string
	TTLSeconds uint
	AutoDelete bool
	Durable    bool
	Lazy       bool
}

// Rabbit -
type Rabbit struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel

	service string
	queues  []Queue
}

// NewRabbit -
func NewRabbit() *Rabbit {
	return &Rabbit{}
}

// Close -
func (mq *Rabbit) Close() error {
	if mq.Channel != nil {
		mq.Channel.Close()
	}
	if mq.Conn != nil {
		mq.Conn.Close()
	}
	return nil
}

// Send -
func (mq *Rabbit) Send(msg IMessage) error {
	queues := msg.GetQueues()
	if len(queues) == 0 {
		return nil
	}
	message, err := msg.MarshalToQueue()
	if err != nil {
		return err
	}
	for _, queue := range queues {
		if err := mq.SendRaw(queue, message); err != nil {
			return err
		}
	}
	return nil
}

// Consume -
func (mq *Rabbit) Consume(queue string) (<-chan amqp.Delivery, error) {
	return mq.Channel.Consume(getQueueName(mq.service, queue), "", false, false, false, false, nil)
}

// SendRaw -
func (mq *Rabbit) SendRaw(queue string, body []byte) error {
	if mq.Channel == nil || mq.Conn == nil {
		return ErrInvalidConnection
	}
	if mq.Conn.IsClosed() {
		return ErrConnectionIsClosed
	}
	return mq.Channel.Publish(
		ChannelNew,
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
}

// GetQueues -
func (mq *Rabbit) GetQueues() []string {
	queues := make([]string, len(mq.queues))
	for i := range mq.queues {
		queues[i] = mq.queues[i].Name
	}
	return queues
}

// RabbitQueueManager -
type RabbitQueueManager struct {
	receiver  IMessageReceiver
	publisher IMessagePublisher
}

// WaitNewRabbit -
func WaitNewRabbit(connection, service string, needPublisher bool, timeout int, queues ...Queue) *RabbitQueueManager {
	var qm *RabbitQueueManager
	var err error

	for qm == nil {
		qm, err = NewQueueManager(connection, service, needPublisher, queues...)
		if err != nil {
			logger.Warning().Msgf("Waiting mq up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}
	return qm
}

// NewQueueManager -
func NewQueueManager(connection, service string, needPublisher bool, queues ...Queue) (*RabbitQueueManager, error) {
	q := RabbitQueueManager{}
	if service != "" && len(queues) > 0 {
		receiver, err := NewRabbitReceiver(connection, service, queues...)
		if err != nil {
			return nil, err
		}
		q.receiver = receiver
	}

	if needPublisher {
		publisher, err := NewRabbitPublisher(connection)
		if err != nil {
			return nil, err
		}
		q.publisher = publisher
	}
	return &q, nil
}

// SendRaw -
func (q RabbitQueueManager) SendRaw(queue string, body []byte) error {
	if q.publisher == nil {
		return nil
	}
	return q.publisher.SendRaw(queue, body)
}

// Send -
func (q RabbitQueueManager) Send(message IMessage) error {
	if q.publisher == nil {
		return nil
	}
	return q.publisher.Send(message)
}

// Consume -
func (q RabbitQueueManager) Consume(queue string) (<-chan amqp.Delivery, error) {
	if q.receiver == nil {
		return nil, nil
	}
	return q.receiver.Consume(queue)
}

// GetQueues -
func (q RabbitQueueManager) GetQueues() []string {
	if q.receiver == nil {
		return nil
	}
	return q.receiver.GetQueues()
}

// Close -
func (q RabbitQueueManager) Close() error {
	if q.publisher != nil {
		q.publisher.Close()
	}
	if q.receiver != nil {
		q.receiver.Close()
	}
	return nil
}

// NewRabbitReceiver -
func NewRabbitReceiver(connection string, service string, queues ...Queue) (*Rabbit, error) {
	mq, err := NewRabbitPublisher(connection)
	if err != nil {
		return nil, err
	}
	mq.queues = queues
	mq.service = service

	for _, queue := range queues {
		args := make(amqp.Table)
		if queue.TTLSeconds > 0 {
			args["x-message-ttl"] = int(queue.TTLSeconds * 1000)
		}
		if queue.Lazy {
			args["x-queue-mode"] = "lazy"
		}

		q, err := mq.Channel.QueueDeclare(getQueueName(service, queue.Name), queue.Durable, queue.AutoDelete, false, false, args)
		if err != nil {
			return nil, err
		}
		if err = mq.Channel.QueueBind(
			q.Name,
			queue.Name,
			ChannelNew,
			false,
			nil,
		); err != nil {
			return nil, err
		}
	}
	return mq, nil
}

// NewRabbitPublisher -
func NewRabbitPublisher(connection string) (*Rabbit, error) {
	mq := NewRabbit()
	conn, err := amqp.Dial(connection)
	if err != nil {
		return nil, err
	}
	mq.Conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	mq.Channel = ch

	err = ch.ExchangeDeclare(
		ChannelNew,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	return mq, err
}
