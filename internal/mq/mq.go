package mq

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

// MQ -
type MQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel

	service string
}

// Close -
func (mq *MQ) Close() {
	if mq.Conn != nil {
		mq.Conn.Close()
	}
	if mq.Channel != nil {
		mq.Channel.Close()
	}
}

// Send -
func (mq *MQ) Send(channel string, queue IQueued, v interface{}) error {
	q := queue.GetQueue()
	if q == "" {
		return nil
	}
	return mq.SendToQueue(channel, q, v)
}

// SendToQueue -
func (mq *MQ) SendToQueue(channel, queue string, v interface{}) error {
	if mq.Channel == nil || mq.Conn == nil {
		return errors.New("Invaid connection or channel")
	}
	if mq.Conn.IsClosed() {
		return errors.New("Connection is closed")
	}
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return mq.Channel.Publish(
		channel,
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
}

// NewReceiver -
func NewReceiver(connection string, queues []string, service string) (*MQ, error) {
	mq := &MQ{
		service: service,
	}
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
	if err != nil {
		return nil, err
	}

	for _, queue := range queues {
		q, err := ch.QueueDeclare(fmt.Sprintf("%s.%s", queue, service), true, false, false, false, nil)
		if err != nil {
			return nil, err
		}
		if err = ch.QueueBind(
			q.Name,
			queue,
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

// Consume -
func (mq *MQ) Consume(queue string) (<-chan amqp.Delivery, error) {
	return mq.Channel.Consume(fmt.Sprintf("%s.%s", queue, mq.service), "", false, false, false, false, nil)
}
