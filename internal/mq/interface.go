package mq

import "github.com/streadway/amqp"

// IMessage -
type IMessage interface {
	GetQueue() string
	Marshal() ([]byte, error)
}

// IMessagePublisher -
type IMessagePublisher interface {
	SendRaw(queue string, body []byte) error
	Send(queue IMessage) error
	Close()
}

// IMessageReceiver -
type IMessageReceiver interface {
	Consume(queue string) (<-chan amqp.Delivery, error)
	GetQueues() []string
	Close()
}
