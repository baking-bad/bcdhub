package main

import (
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/streadway/amqp"
)

// BulkHandler -
type BulkHandler func(ids []string) error

// BulkManager -
type BulkManager struct {
	timeout time.Duration
	queue   []amqp.Delivery
	handler BulkHandler

	ticker *time.Ticker
	stop   chan struct{}

	lock sync.Mutex
	wg   sync.WaitGroup
}

// NewBulkManager -
func NewBulkManager(capacity, timeout int, handler BulkHandler) *BulkManager {
	return &BulkManager{
		timeout: time.Duration(timeout) * time.Second,
		queue:   make([]amqp.Delivery, 0, capacity),
		stop:    make(chan struct{}),
		ticker:  time.NewTicker(time.Duration(timeout) * time.Second),
		handler: handler,
	}
}

// Add -
func (bm *BulkManager) Add(data amqp.Delivery) {
	defer bm.lock.Unlock()
	bm.lock.Lock()
	{
		length := len(bm.queue)
		if length < cap(bm.queue) {
			bm.queue = append(bm.queue, data)
		}
		if len(bm.queue) == cap(bm.queue) {
			bm.ticker.Stop()
			bm.process()
			bm.ticker = time.NewTicker(bm.timeout)
		}
	}
}

// Run -
func (bm *BulkManager) Run() {
	defer bm.wg.Done()

	bm.wg.Add(1)
	bm.ticker = time.NewTicker(bm.timeout)
	defer bm.ticker.Stop()

	for {
		select {
		case <-bm.stop:
			return
		case <-bm.ticker.C:
			bm.lock.Lock()
			{
				bm.process()
				bm.lock.Unlock()
			}
		}
	}
}

func (bm *BulkManager) process() {
	for len(bm.queue) > 0 {
		ids := make([]string, len(bm.queue))
		for i := range bm.queue {
			ids[i] = parseID(bm.queue[i].Body)
		}
		if err := bm.handler(ids); err != nil {
			logger.Error(err)
			continue
		}
		for i := range bm.queue {
			if err := bm.queue[i].Ack(false); err != nil {
				logger.Errorf("Error acknowledging message: %s", err)
				return
			}
		}
		bm.queue = make([]amqp.Delivery, 0, cap(bm.queue))
	}
}

// Stop -
func (bm *BulkManager) Stop() {
	bm.stop <- struct{}{}
	bm.wg.Wait()

	close(bm.stop)
}
