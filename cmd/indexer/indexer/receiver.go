package indexer

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Block -
type Block struct {
	Header noderpc.Header
	OPG    []noderpc.LightOperationGroup
}

// Receiver -
type Receiver struct {
	rpc    noderpc.INode
	queue  chan int64
	failed chan int64
	blocks chan *Block

	threads chan struct{}
	present map[int64]struct{}
	mx      sync.RWMutex
	wg      sync.WaitGroup
}

// NewReceiver -
func NewReceiver(rpc noderpc.INode, queueSize, threadsCount int64) *Receiver {
	if queueSize == 0 || queueSize > 100 {
		queueSize = 10
	}
	if threadsCount == 0 {
		threadsCount = 2
	}
	return &Receiver{
		rpc:     rpc,
		queue:   make(chan int64, queueSize),
		failed:  make(chan int64, queueSize),
		blocks:  make(chan *Block, queueSize),
		threads: make(chan struct{}, threadsCount),
		present: make(map[int64]struct{}),
	}
}

// AddTask -
func (r *Receiver) AddTask(level int64) {
	r.mx.RLock()
	if _, ok := r.present[level]; ok {
		r.mx.RUnlock()
		return
	}
	r.mx.RUnlock()

	r.mx.Lock()
	{
		r.present[level] = struct{}{}
		r.queue <- level
	}
	r.mx.Unlock()
}

// Start -
func (r *Receiver) Start(ctx context.Context) {
	go r.start(ctx)
}

// Blocks -
func (r *Receiver) Blocks() <-chan *Block {
	return r.blocks
}

func (r *Receiver) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			r.wg.Wait()
			close(r.threads)
			close(r.blocks)
			close(r.failed)
			close(r.queue)
			return
		case level := <-r.failed:
			r.job(level)
		case level := <-r.queue:
			r.job(level)
		}
	}
}

func (r *Receiver) get(level int64) (Block, error) {
	header, err := r.rpc.GetHeader(level)
	if err != nil {
		return Block{}, err
	}
	if level < 2 {
		return Block{header, make([]noderpc.LightOperationGroup, 0)}, nil
	}
	opg, err := r.rpc.GetLightOPG(level)
	if err != nil {
		return Block{}, err
	}
	return Block{header, opg}, nil
}

func (r *Receiver) job(level int64) {
	r.threads <- struct{}{}
	r.wg.Add(1)
	go func() {
		defer func() {
			<-r.threads
			r.wg.Done()
		}()

		block, err := r.get(level)
		if err != nil {
			logger.Error().Int64("block", level).Err(err).Msg("Receiver.get")
			r.failed <- level
			return
		}
		r.blocks <- &block

		r.mx.Lock()
		{
			delete(r.present, level)
		}
		r.mx.Unlock()
	}()
}
