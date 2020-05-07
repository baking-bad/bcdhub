package ws

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/baking-bad/bcdhub/cmd/api/ws/channels"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/gorilla/websocket"
	"github.com/valyala/fastjson"
)

// ClientHandler -
type ClientHandler func(*Client, []byte) error

// ClientEvent -
type ClientEvent func([]byte) error

// Client -
type Client struct {
	id   int
	conn *websocket.Conn

	sender chan channels.Message
	stop   chan struct{}

	subscriptions map[string]struct{}
	mux           sync.Mutex

	handlers map[string]ClientHandler
	sendMux  sync.Mutex

	onSubscribe   ClientEvent
	onUnsubscribe ClientEvent

	hub *Hub
}

// NewClient -
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		id:   rand.Int(),
		conn: conn,

		sender: make(chan channels.Message),
		stop:   make(chan struct{}),

		subscriptions: make(map[string]struct{}),
		handlers:      make(map[string]ClientHandler),
	}
}

// Send -
func (c *Client) Send(msg channels.Message) {
	c.mux.Lock()
	if _, ok := c.subscriptions[msg.ChannelName]; ok {
		c.sender <- msg
	}
	c.mux.Unlock()
}

// Run -
func (c *Client) Run() {
	go c.send()
	go c.receive()
}

// Close -
func (c *Client) Close() {
	close(c.stop)
	close(c.sender)
	c.conn.Close()
	c.hub.RemoveClient(c)
}

// AddHandler -
func (c *Client) AddHandler(name string, handler ClientHandler) {
	c.handlers[name] = handler
}

func (c *Client) sendMessage(message interface{}) error {
	c.sendMux.Lock()
	defer c.sendMux.Unlock()

	return c.conn.WriteJSON(message)
}

func (c *Client) sendError(err error) {
	msg := StatusMessage{
		Status: ErrorStatus,
		Text:   err.Error(),
	}
	if err := c.sendMessage(msg); err != nil {
		logger.Error(err)
	}
}

func (c *Client) sendOk() error {
	msg := StatusMessage{
		Status: OkStatus,
	}
	return c.sendMessage(msg)
}

func (c *Client) receive() {
	var p fastjson.Parser
	for {
		select {
		case <-c.stop:
			return
		default:
			messageType, data, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) {
					c.Close()
				} else {
					logger.Error(err)
				}
				continue
			}

			switch messageType {
			case websocket.TextMessage:
				val, err := p.ParseBytes(data)
				if err != nil {
					logger.Error(err)
					continue
				}
				action := string(val.GetStringBytes("action"))
				handler, ok := c.handlers[action]
				if !ok {
					c.sendError(fmt.Errorf("Unknown handler action: %s", action))
					continue
				}
				if err := handler(c, data); err != nil {
					c.sendError(err)
					continue
				}
			}
		}
	}
}

func (c *Client) send() {
	for {
		select {
		case <-c.stop:
			return
		case msg := <-c.sender:
			if msg.ChannelName != "" && msg.Body != nil {
				if err := c.sendMessage(msg); err != nil {
					c.sendError(err)
				}
			}
		}
	}
}
