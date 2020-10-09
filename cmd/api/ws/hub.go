package ws

import (
	"fmt"
	"sync"

	"github.com/baking-bad/bcdhub/cmd/api/ws/channels"
	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

// Hub -
type Hub struct {
	sources []datasources.DataSource
	clients sync.Map
	public  sync.Map

	elastic elastic.IElastic

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewHub -
func NewHub(opts ...HubOption) *Hub {
	h := &Hub{
		sources: make([]datasources.DataSource, 0),

		stop: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// DefaultHub -
func DefaultHub(connectionElastic string, timeoutElastic int, messageQueue *mq.QueueManager) *Hub {
	es := elastic.WaitNew([]string{connectionElastic}, timeoutElastic)
	hub := NewHub(
		WithRabbitSource(messageQueue),
		WithElastic(es),
	)

	hub.AddPublicChannel(channels.NewStatsChannel(
		channels.WithSource(hub.sources, datasources.RabbitType),
		channels.WithElasticSearch(es),
	))
	return hub
}

// AddPublicChannel -
func (h *Hub) AddPublicChannel(channel channels.Channel) {
	h.public.Store(channel.GetName(), channel)
}

// AddClient -
func (h *Hub) AddClient(client *Client) {
	client.hub = h
	client.AddHandler("subscribe", subscribeHandler)
	client.AddHandler("unsubscribe", unsubscribeHandler)
	h.clients.Store(client.id, client)
}

// GetPublicChannel -
func (h *Hub) GetPublicChannel(name string) (channels.Channel, bool) {
	c, ok := h.public.Load(name)
	if !ok {
		return nil, ok
	}
	channel, ok := c.(channels.Channel)
	return channel, ok
}

// RemoveClient -
func (h *Hub) RemoveClient(client *Client) {
	if _, ok := h.clients.Load(client.id); ok {
		h.clients.Delete(client.id)
	}
}

// Run -
func (h *Hub) Run() {
	for i := range h.sources {
		h.sources[i].Run()
	}

	h.public.Range(func(key, val interface{}) bool {
		ch := val.(channels.Channel)
		h.runChannel(ch)
		return true
	})
}

func (h *Hub) runChannel(channel channels.Channel) {
	h.wg.Add(1)
	go h.listenChannel(channel)
	channel.Run()
}

// Stop -
func (h *Hub) Stop() {
	defer h.wg.Wait()

	close(h.stop)

	h.clients.Range(func(key, val interface{}) bool {
		client := val.(*Client)
		client.Close()
		return true
	})

	h.public.Range(func(key, val interface{}) bool {
		ch := val.(channels.Channel)
		ch.Stop()
		return true
	})
}

func (h *Hub) listenChannel(channel channels.Channel) {
	defer h.wg.Done()
	for {
		select {
		case <-h.stop:
			return
		case msg := <-channel.Listen():
			if msg.Body == nil && msg.ChannelName == "" {
				return
			}
			h.clients.Range(func(key, val interface{}) bool {
				client := val.(*Client)
				client.Send(msg)
				return true
			})
		}
	}
}

func createDynamicChannels(c *Client, channelName string, data *fastjson.Value) (channels.Channel, error) {
	switch channelName {
	case "operations":
		address := parseString(data, "address")
		network := parseString(data, "network")

		operationsChannelName := fmt.Sprintf("%s_%s_%s", channelName, network, address)
		if _, ok := c.subscriptions.Load(operationsChannelName); ok {
			return nil, nil
		}
		return channels.NewOperationsChannel(address, network,
			channels.WithSource(c.hub.sources, datasources.RabbitType),
			channels.WithElasticSearch(c.hub.elastic),
		), nil
	default:
		return nil, errors.Errorf("Unknown channel: %s", channelName)
	}
}

func subscribeHandler(c *Client, data []byte) error {
	var p fastjson.Parser
	val, err := p.ParseBytes(data)
	if err != nil {
		return err
	}
	channelName := parseString(val, "channel")
	channel, ok := c.hub.GetPublicChannel(channelName)
	if !ok {
		channel, err = createDynamicChannels(c, channelName, val)
		if err != nil {
			return err
		}
		if channel == nil {
			return nil
		}
		c.hub.runChannel(channel)
	}
	c.subscriptions.Store(channel.GetName(), channel)

	return channel.Init()
}

func unsubscribeHandler(c *Client, data []byte) error {
	var p fastjson.Parser
	val, err := p.ParseBytes(data)
	if err != nil {
		return err
	}
	channelName := parseString(val, "channel")
	if channel, ok := c.GetSubscription(channelName); ok {
		if _, isPublic := c.hub.public.Load(channelName); !isPublic {
			channel.Stop()
		}
	}
	c.subscriptions.Delete(channelName)
	return c.sendOk(fmt.Sprintf("unsubscribed from %s", channelName))
}

func parseString(val *fastjson.Value, key string) string {
	return string(val.GetStringBytes(key))
}
