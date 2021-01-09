package datasources

import (
	"sync"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/pkg/errors"
)

var rabbitStopped = errors.Errorf("WS_RABBIT_STOPPED")

// RabbitMQ -
type RabbitMQ struct {
	*DefaultSource

	source mq.Mediator

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewRabbitMQ -
func NewRabbitMQ(messageQueue mq.Mediator) (*RabbitMQ, error) {
	return &RabbitMQ{
		DefaultSource: NewDefaultSource(),
		source:        messageQueue,
		stop:          make(chan struct{}),
	}, nil
}

// Run -
func (c *RabbitMQ) Run() {
	if len(c.source.GetQueues()) == 0 {
		logger.Warning("Empty rabbit queues")
		return
	}

	for _, queue := range c.source.GetQueues() {
		c.wg.Add(1)
		go c.listenChannel(queue)
	}
}

// Stop -
func (c *RabbitMQ) Stop() {
	close(c.stop)
	c.wg.Wait()
	c.subscribers.Range(func(key, val interface{}) bool {
		close(key.(chan Data))
		return true
	})
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
				if errors.Is(err, rabbitStopped) {
					return
				}
				logger.Errorf("[%s data source] %s", c.GetType(), err.Error())
			}
		}
	}
}

func (c *RabbitMQ) handler(data mq.Data) error {
	switch data.GetKey() {
	case mq.QueueOperations, mq.QueueBlocks:
		val := Data{
			Type: c.GetType(),
			Kind: data.GetKey(),
			Body: data.GetBody(),
		}

		c.subscribers.Range(func(key, value interface{}) bool {
			ch := key.(chan Data)
			ch <- val
			return true
		})
	default:
		if data.GetKey() == "" {
			logger.Warning("Rabbit MQ server stopped! API need to be restarted. Closing connection...")
			return rabbitStopped
		}
		return errors.Errorf("Unknown data routing key %s", data.GetKey())
	}

	if err := data.Ack(false); err != nil {
		return errors.Errorf("Error acknowledging message: %s", err)
	}
	return nil
}
