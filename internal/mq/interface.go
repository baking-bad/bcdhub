package mq

import (
	"io"

	"github.com/streadway/amqp"
)

// IMessage -
type IMessage interface {
	GetQueues() []string
	MarshalToQueue() ([]byte, error)
}

// Publisher -
type Publisher interface {
	SendRaw(queue string, body []byte) error
	Send(queue IMessage) error
}

// IMessagePublisher -
type IMessagePublisher interface {
	Publisher
	io.Closer
}

// Receiver -
type Receiver interface {
	Consume(queue string) (<-chan amqp.Delivery, error)
	GetQueues() []string
}

// IMessageReceiver -
type IMessageReceiver interface {
	Receiver
	io.Closer
}

// Mediator -
type Mediator interface {
	Publisher
	Receiver
	io.Closer
}
