package main

import (
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
)

// BulkHandler -
type BulkHandler func(ids []string) error

// BulkManager -
type BulkManager struct {
	timeout time.Duration
	queue   []mq.Data
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
		queue:   make([]mq.Data, 0, capacity),
		stop:    make(chan struct{}),
		ticker:  time.NewTicker(time.Duration(timeout) * time.Second),
		handler: handler,
	}
}

// Add -
func (bm *BulkManager) Add(data mq.Data) {
	defer bm.lock.Unlock()
	bm.lock.Lock()
	{
		if len(bm.queue) == cap(bm.queue) {
			bm.process()
		}
		bm.queue = append(bm.queue, data)
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
	if len(bm.queue) == 0 {
		return
	}

	ids := make([]string, len(bm.queue))
	for i := range bm.queue {
		ids[i] = parseID(bm.queue[i].GetBody())
	}
	if err := bm.handler(ids); err != nil {
		logger.Error(err)
		return
	}
	for i := range bm.queue {
		if err := bm.queue[i].Ack(false); err != nil {
			logger.Errorf("Error acknowledging message: %s", err)
			return
		}
	}
	bm.queue = make([]mq.Data, 0, cap(bm.queue))
	bm.ticker.Stop()
	bm.ticker = time.NewTicker(bm.timeout)
}

// Stop -
func (bm *BulkManager) Stop() {
	bm.stop <- struct{}{}
	bm.wg.Wait()

	close(bm.stop)
}
