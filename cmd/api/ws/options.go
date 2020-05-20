package ws

import (
	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
	"github.com/baking-bad/bcdhub/internal/logger"
)

// HubOption -
type HubOption func(*Hub)

// WithSource -
func WithSource(source datasources.DataSource) HubOption {
	return func(h *Hub) {
		h.sources = append(h.sources, source)
	}
}

// WithRabbitSource -
func WithRabbitSource(connection string, queues []string) HubOption {
	return func(h *Hub) {
		rmq, err := datasources.NewRabbitMQ(connection, queues)
		if err != nil {
			logger.Error(err)
			return
		}
		h.sources = append(h.sources, rmq)
	}
}
