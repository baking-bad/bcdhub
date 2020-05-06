package ws

import "github.com/baking-bad/bcdhub/cmd/api/ws/channels"

// Client -
type Client struct {
	messages chan struct{}
}

// NewClient -
func NewClient() *Client {
	return &Client{
		messages: make(chan struct{}),
	}
}

// Close -
func (c *Client) Close() {
	close(c.messages)
}

// Send -
func (c *Client) Send(msg channels.Message) error {
	return nil
}
