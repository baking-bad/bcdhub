package datasources

import (
	"sync"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// RabbitMQ -
type RabbitMQ struct {
	*DefaultSource

	source *mq.MQ
	queues []string

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewRabbitMQ -
func NewRabbitMQ(connectionString string, queues []string) (*RabbitMQ, error) {
	messageQueue, err := mq.NewReceiver(connectionString, queues, "ws")
	if err != nil {
		return nil, err
	}
	return &RabbitMQ{
		DefaultSource: NewDefaultSource(),
		source:        messageQueue,
		queues:        queues,
		stop:          make(chan struct{}),
	}, nil
}

// Run -
func (c *RabbitMQ) Run() {
	if len(c.queues) == 0 {
		logger.Warning("Empty rabbit queues")
		return
	}

	for i := range c.queues {
		c.wg.Add(1)
		go c.listenChannel(c.queues[i])
	}
}

// Stop -
func (c *RabbitMQ) Stop() {
	close(c.stop)
	c.wg.Wait()
	for ch := range c.subscribers {
		close(ch)
	}
	c.source.Close()
}

// GetType -
func (c *RabbitMQ) GetType() string {
	return RabbitType
}

func (c *RabbitMQ) listenChannel(queue string) {
	defer c.wg.Done()

	msgs, err := c.source.Consume(queue)
	if err != nil {
		logger.Errorf("[%s data source] %s. Stop.", c.GetType(), err.Error())
		return
	}

	for {
		select {
		case <-c.stop:
			return
		case msg := <-msgs:
			if err := c.handler(msg); err != nil {
				if err.Error() == "WS_RABBIT_STOPPED" {
					return
				}
				logger.Errorf("[%s data source] %s", c.GetType(), err.Error())
			}
		}
	}
}

func (c *RabbitMQ) handler(data amqp.Delivery) error {
	switch data.RoutingKey {
	case mq.QueueOperations:
		val := Data{
			Type: c.GetType(),
			Body: data.Body,
		}
		for ch := range c.subscribers {
			ch <- val
		}
	default:
		if data.RoutingKey == "" {
			logger.Warning("Rabbit MQ server stopped! API need to be restarted. Closing connection...")
			return errors.Errorf("WS_RABBIT_STOPPED")
		}
		return errors.Errorf("Unknown data routing key %s", data.RoutingKey)
	}

	if err := data.Ack(false); err != nil {
		return errors.Errorf("Error acknowledging message: %s", err)
	}
	return nil
}
