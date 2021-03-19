package mq

import "errors"

// Channels
const (
	ChannelNew = "new"
)

// Queues
const (
	QueueContracts    = "contracts"
	QueueOperations   = "operations"
	QueueCompilations = "compilations"
	QueueBigMapDiffs  = "bigmapdiffs"
	QueueBlocks       = "blocks"
)

// URL Prefixes
const (
	RabbitURLPrefix = "amqp"
	NatsURLPrefix   = "nats"
)

// Errors
var (
	ErrUnknownQueue       = errors.New("Unknown queue name")
	ErrConnectionIsClosed = errors.New("Connection is closed")
	ErrInvalidConnection  = errors.New("Invalid connection or channel")
)
