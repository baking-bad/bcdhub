package ws

import (
	"fmt"
	"sync"

	"github.com/baking-bad/bcdhub/cmd/api/ws/channels"
	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
	"github.com/valyala/fastjson"
)

// Hub -
type Hub struct {
	sources []datasources.DataSource
	clients map[int]*Client
	public  map[string]channels.Channel

	stop chan struct{}
	wg   sync.WaitGroup

	mux sync.Mutex
}

// NewHub -
func NewHub(opts ...HubOption) *Hub {
	h := &Hub{
		sources: make([]datasources.DataSource, 0),
		clients: make(map[int]*Client),
		public:  make(map[string]channels.Channel),

		stop: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// DefaultHub -
func DefaultHub(elasticConnection string, elasticTimeout int, rabbitConnection string, queues []string) *Hub {
	hub := NewHub(
		WithRabbitSource(rabbitConnection, queues),
	)
	hub.AddPublicChannel(channels.NewStatsChannel(
		channels.WithSource(hub.sources, datasources.RabbitType),
		channels.WithElasticSearch(elasticConnection, elasticTimeout),
	))
	return hub
}

// AddPublicChannel -
func (h *Hub) AddPublicChannel(channel channels.Channel) {
	h.public[channel.GetName()] = channel
}

// AddClient -
func (h *Hub) AddClient(client *Client) {
	client.hub = h
	client.AddHandler("subscribe", subscribeHandler)
	client.AddHandler("unsubscribe", unsubscribeHandler)
	h.mux.Lock()
	h.clients[client.id] = client
	h.mux.Unlock()
}

// RemoveClient -
func (h *Hub) RemoveClient(client *Client) {
	if _, ok := h.clients[client.id]; ok {
		h.mux.Lock()
		delete(h.clients, client.id)
		h.mux.Unlock()
	}
}

// Run -
func (h *Hub) Run() {
	for i := range h.sources {
		h.sources[i].Run()
	}

	for _, channel := range h.public {
		h.wg.Add(1)
		go h.listenChannel(channel)
		channel.Run()
	}
}

// Stop -
func (h *Hub) Stop() {
	close(h.stop)
	for _, channel := range h.public {
		channel.Stop()
	}
	for _, client := range h.clients {
		client.Close()
	}
	h.wg.Wait()
}

func (h *Hub) listenChannel(channel channels.Channel) {
	defer h.wg.Done()
	for {
		select {
		case <-h.stop:
			return
		case msg := <-channel.Listen():
			for _, client := range h.clients {
				client.Send(msg)
			}
		}
	}
}

func subscribeHandler(c *Client, data []byte) error {
	var p fastjson.Parser
	val, err := p.ParseBytes(data)
	if err != nil {
		return err
	}
	channelName := string(val.GetStringBytes("channel"))
	channel, ok := c.hub.public[channelName]
	if !ok {
		return fmt.Errorf("Unknown channel: %s", channelName)
	}
	c.mux.Lock()
	c.subscriptions[channelName] = struct{}{}
	c.mux.Unlock()

	return channel.Init()
}

func unsubscribeHandler(c *Client, data []byte) error {
	var p fastjson.Parser
	val, err := p.ParseBytes(data)
	if err != nil {
		return err
	}
	channelName := string(val.GetStringBytes("channel"))
	if _, ok := c.hub.public[channelName]; !ok {
		return fmt.Errorf("Unknown channel: %s", channelName)
	}
	c.mux.Lock()
	delete(c.subscriptions, channelName)
	c.mux.Unlock()
	return c.sendOk()
}
