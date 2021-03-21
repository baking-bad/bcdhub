package main

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/streadway/amqp"
)

// BulkHandler -
type BulkHandler func(ids []int64) error

// BulkManager -
type BulkManager struct {
	queue   []amqp.Delivery
	handler BulkHandler

	lastAction time.Time
	timeout    time.Duration
}

// NewBulkManager -
func NewBulkManager(capacity, timeout int, handler BulkHandler) *BulkManager {
	return &BulkManager{
		queue:   make([]amqp.Delivery, 0, capacity),
		handler: handler,
		timeout: time.Duration(timeout) * time.Second,
	}
}

// Add -
func (bm *BulkManager) Add(data amqp.Delivery) {
	force := bm.lastAction.IsZero() || time.Since(bm.lastAction) > bm.timeout
	if bm.process(force) {
		bm.queue = append(bm.queue, data)
	}
}

func (bm *BulkManager) process(force bool) bool {
	execute := len(bm.queue) == cap(bm.queue) || (force && len(bm.queue) > 0)
	if !execute {
		return true
	}

	defer func() {
		bm.lastAction = time.Now()
	}()

	ids := make([]int64, len(bm.queue))
	for i := range bm.queue {
		id, err := parseID(bm.queue[i].Body)
		if err != nil {
			logger.Error(err)
			continue
		}
		ids[i] = id
	}
	if err := bm.handler(ids); err != nil {
		logger.Error(err)
		return false
	}
	for i := range bm.queue {
		if err := bm.queue[i].Ack(false); err != nil {
			logger.Errorf("Error acknowledging message: %s", err)
			return false
		}
	}
	bm.queue = make([]amqp.Delivery, 0, cap(bm.queue))

	return true
}
