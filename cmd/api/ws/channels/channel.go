package channels

import (
	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
)

// Channel -
type Channel interface {
	GetName() string
	Run()
	Listen() <-chan Message
	Stop()
	Init() error
}

// DefaultChannel -
type DefaultChannel struct {
	sources []datasources.DataSource

	ctx *handlers.Context
}

// NewDefaultChannel -
func NewDefaultChannel(opts ...ChannelOption) *DefaultChannel {
	c := &DefaultChannel{
		sources: make([]datasources.DataSource, 0),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Message -
type Message struct {
	ChannelName string      `json:"channel_name"`
	Body        interface{} `json:"body"`
}
