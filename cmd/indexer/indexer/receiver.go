package indexer

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/protocols"
)

// Block -
type Block struct {
	Header   noderpc.Header
	OPG      []noderpc.LightOperationGroup
	Metadata *noderpc.Metadata
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
	}
	r.mx.Unlock()
	r.queue <- level
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
			r.job(ctx, level)
		case level := <-r.queue:
			r.job(ctx, level)
		}
	}
}

func (r *Receiver) get(ctx context.Context, level int64) (Block, error) {
	var block Block
	header, err := r.rpc.GetHeader(ctx, level)
	if err != nil {
		return block, err
	}
	block.Header = header

	if level < 2 {
		block.OPG = make([]noderpc.LightOperationGroup, 0)
		return block, nil
	}

	opg, err := r.rpc.GetLightOPG(ctx, level)
	if err != nil {
		return block, err
	}
	block.OPG = opg

	if protocols.NeedImplicitParsing(header.Hash) {
		metadata, err := r.rpc.GetBlockMetadata(ctx, level)
		if err != nil {
			return block, err
		}
		block.Metadata = &metadata
	}

	return block, nil
}

func (r *Receiver) job(ctx context.Context, level int64) {
	r.threads <- struct{}{}
	r.wg.Add(1)
	go func() {
		defer func() {
			<-r.threads
			r.wg.Done()
		}()

		block, err := r.get(ctx, level)
		if err != nil {
			if ctx.Err() == nil {
				logger.Error().Int64("block", level).Err(err).Msg("Receiver.get")
				r.failed <- level
			}
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
