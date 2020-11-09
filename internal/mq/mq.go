package mq

import (
	"errors"
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/streadway/amqp"
)

// Queue -
type Queue struct {
	Name       string
	AutoDelete bool
	Durable    bool
}

// MQ -
type MQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel

	service string
	queues  []Queue
}

// Close -
func (mq *MQ) Close() error {
	if mq.Conn != nil {
		mq.Conn.Close()
	}
	if mq.Channel != nil {
		mq.Channel.Close()
	}
	return nil
}

// Send -
func (mq *MQ) Send(msg IMessage) error {
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
func (mq *MQ) Consume(queue string) (<-chan amqp.Delivery, error) {
	return mq.Channel.Consume(fmt.Sprintf("%s.%s", queue, mq.service), "", false, false, false, false, nil)
}

// SendRaw -
func (mq *MQ) SendRaw(queue string, body []byte) error {
	if mq.Channel == nil || mq.Conn == nil {
		return errors.New("Invaid connection or channel")
	}
	if mq.Conn.IsClosed() {
		return errors.New("Connection is closed")
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
func (mq *MQ) GetQueues() []string {
	queues := make([]string, len(mq.queues))
	for i := range mq.queues {
		queues[i] = mq.queues[i].Name
	}
	return queues
}

// QueueManager -
type QueueManager struct {
	receiver  IMessageReceiver
	publisher IMessagePublisher
}

// WaitNew -
func WaitNew(connection, service string, needPublisher bool, timeout int, queues ...Queue) *QueueManager {
	var qm *QueueManager
	var err error

	for qm == nil {
		qm, err = NewQueueManager(connection, service, needPublisher, queues...)
		if err != nil {
			logger.Warning("Waiting mq up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}
	return qm
}

// NewQueueManager -
func NewQueueManager(connection, service string, needPublisher bool, queues ...Queue) (*QueueManager, error) {
	q := QueueManager{}
	if service != "" && len(queues) > 0 {
		receiver, err := NewReceiver(connection, service, queues...)
		if err != nil {
			return nil, err
		}
		q.receiver = receiver
	}

	if needPublisher {
		publisher, err := NewPublisher(connection)
		if err != nil {
			return nil, err
		}
		q.publisher = publisher
	}
	return &q, nil
}

// SendRaw -
func (q QueueManager) SendRaw(queue string, body []byte) error {
	if q.publisher == nil {
		return nil
	}
	return q.publisher.SendRaw(queue, body)
}

// Send -
func (q QueueManager) Send(message IMessage) error {
	if q.publisher == nil {
		return nil
	}
	return q.publisher.Send(message)
}

// Consume -
func (q QueueManager) Consume(queue string) (<-chan amqp.Delivery, error) {
	if q.receiver == nil {
		return nil, nil
	}
	return q.receiver.Consume(queue)
}

// GetQueues -
func (q QueueManager) GetQueues() []string {
	if q.receiver == nil {
		return nil
	}
	return q.receiver.GetQueues()
}

// Close -
func (q QueueManager) Close() error {
	if q.publisher != nil {
		q.publisher.Close()
	}
	if q.receiver != nil {
		q.receiver.Close()
	}
	return nil
}

// NewReceiver -
func NewReceiver(connection string, service string, queues ...Queue) (*MQ, error) {
	mq, err := NewPublisher(connection)
	if err != nil {
		return nil, err
	}
	mq.queues = queues
	mq.service = service

	for _, queue := range queues {
		q, err := mq.Channel.QueueDeclare(fmt.Sprintf("%s.%s", queue.Name, service), queue.Durable, queue.AutoDelete, false, false, nil)
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

// NewPublisher -
func NewPublisher(connection string) (*MQ, error) {
	mq := &MQ{}
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
