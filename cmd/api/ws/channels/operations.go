package channels

import (
	"fmt"
	"sync"

	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

// OperationsChannel -
type OperationsChannel struct {
	*DefaultChannel
	Address string
	Network string

	messages chan Message
	stop     chan struct{}
	wg       sync.WaitGroup
}

// NewOperationsChannel -
func NewOperationsChannel(address, network string, opts ...ChannelOption) *OperationsChannel {
	return &OperationsChannel{
		DefaultChannel: NewDefaultChannel(opts...),
		Address:        address,
		Network:        network,

		messages: make(chan Message, 10),
		stop:     make(chan struct{}),
	}
}

// GetName -
func (c *OperationsChannel) GetName() string {
	return fmt.Sprintf("operations_%s_%s", c.Network, c.Address)
}

// Run -
func (c *OperationsChannel) Run() {
	if len(c.sources) == 0 {
		logger.Errorf("[%s] Empty source list", c.GetName())
		return
	}

	for i := range c.sources {
		c.wg.Add(1)
		go c.listen(c.sources[i])
	}
}

// Listen -
func (c *OperationsChannel) Listen() <-chan Message {
	return c.messages
}

// Stop -
func (c *OperationsChannel) Stop() {
	close(c.stop)
	c.wg.Wait()
	close(c.messages)
}

// Init -
func (c *OperationsChannel) Init() error {
	c.messages <- Message{
		ChannelName: c.GetName(),
		Body:        "ok",
	}
	return nil
}

func (c *OperationsChannel) listen(source datasources.DataSource) {
	defer c.wg.Done()

	ch := source.Subscribe()
	for {
		select {
		case <-c.stop:
			source.Unsubscribe(ch)
			return
		case data := <-ch:
			if data.Type != datasources.RabbitType {
				continue
			}
			if err := c.createMessage(data); err != nil {
				logger.Error(err)
			}
		}
	}
}

func (c *OperationsChannel) createMessage(data datasources.Data) error {
	op := models.Operation{ID: string(data.Body.([]byte))}
	if err := c.es.GetByID(&op); err != nil {
		return errors.Errorf("[OperationsChannel.createMessage] Find operation error: %s", err)
	}
	if op.Network != c.Network {
		return nil
	}
	if op.Destination != c.Address && op.Source != c.Address {
		return nil
	}
	operations, err := c.es.GetOperations(
		map[string]interface{}{
			"hash": op.Hash,
		},
		0,
		true,
	)
	if err != nil && !elastic.IsRecordNotFound(err) {
		return err
	}

	response, err := handlers.PrepareOperations(c.es, operations, true)
	if err != nil {
		return err
	}
	c.messages <- Message{
		ChannelName: c.GetName(),
		Body:        response,
	}
	return nil
}
