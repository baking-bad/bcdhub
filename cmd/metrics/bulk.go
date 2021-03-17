package main

import (
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
)

// BulkHandler -
type BulkHandler func(ids []int64) error

// BulkManager -
type BulkManager struct {
	queue   []mq.Data
	handler BulkHandler

	ticker *ticker
	stop   chan struct{}

	lock sync.Mutex
	wg   sync.WaitGroup
}

// NewBulkManager -
func NewBulkManager(capacity, timeout int, handler BulkHandler) *BulkManager {
	return &BulkManager{
		queue:   make([]mq.Data, 0, capacity),
		stop:    make(chan struct{}),
		ticker:  newTicker(timeout),
		handler: handler,
	}
}

// Add -
func (bm *BulkManager) Add(data mq.Data) {
	defer bm.lock.Unlock()
	bm.lock.Lock()
	{
		if bm.process(false) {
			bm.queue = append(bm.queue, data)
		}
	}
}

// Run -
func (bm *BulkManager) Run() {
	defer bm.wg.Done()

	bm.wg.Add(1)

	for {
		select {
		case <-bm.stop:
			return
		case <-bm.ticker.listen():
			bm.lock.Lock()
			{
				bm.process(true)
			}
			bm.lock.Unlock()
		}
	}
}

func (bm *BulkManager) process(force bool) bool {
	if len(bm.queue) != cap(bm.queue) && !(force && len(bm.queue) > 0) {
		return true
	}
	bm.ticker.stop()

	ids := make([]int64, len(bm.queue))
	for i := range bm.queue {
		id, err := parseID(bm.queue[i].GetBody())
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
	bm.queue = make([]mq.Data, 0, cap(bm.queue))
	bm.ticker.start()
	return true
}

// Stop -
func (bm *BulkManager) Stop() {
	bm.stop <- struct{}{}
	bm.wg.Wait()

	bm.ticker.stop()
	close(bm.stop)
}

type ticker struct {
	period time.Duration
	ticker time.Ticker
}

func newTicker(timeout int) *ticker {
	period := time.Duration(timeout) * time.Second
	return &ticker{period, *time.NewTicker(period)}
}

func (t *ticker) start() {
	t.ticker = *time.NewTicker(t.period)
}

func (t *ticker) listen() <-chan time.Time {
	return t.ticker.C
}

func (t *ticker) stop() {
	t.ticker.Stop()
}
