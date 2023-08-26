package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/dipdup-io/workerpool"
	"github.com/pkg/errors"
)

// Block -
type Block struct {
	Header   noderpc.Header
	OPG      []noderpc.LightOperationGroup
	Metadata *noderpc.Metadata
}

// Receiver -
type Receiver struct {
	rpc       noderpc.INode
	blocks    chan *Block
	pool      *workerpool.Pool[int64]
	inProcess Map[int64, struct{}]
}

// NewReceiver -
func NewReceiver(rpc noderpc.INode, queueSize, threadsCount int64) *Receiver {
	if queueSize == 0 || queueSize > 100 {
		queueSize = 10
	}
	if threadsCount == 0 {
		threadsCount = 2
	}
	receiver := &Receiver{
		rpc:       rpc,
		blocks:    make(chan *Block, queueSize),
		inProcess: NewMap[int64, struct{}](),
	}
	receiver.pool = workerpool.NewPool(receiver.job, int(threadsCount))
	return receiver
}

// AddTask -
func (r *Receiver) AddTask(level int64) {
	if exists := r.inProcess.Exists(level); exists {
		return
	}
	r.inProcess.Set(level, struct{}{})
	r.pool.AddTask(level)
}

// Start -
func (r *Receiver) Start(ctx context.Context) {
	r.pool.Start(ctx)
}

// Close -
func (r *Receiver) Close() error {
	if err := r.pool.Close(); err != nil {
		return err
	}

	close(r.blocks)
	return nil
}

// Blocks -
func (r *Receiver) Blocks() <-chan *Block {
	return r.blocks
}

func (r *Receiver) get(ctx context.Context, level int64) (Block, error) {
	var block Block
	header, err := r.rpc.Block(ctx, level)
	if err != nil {
		return block, err
	}
	block.Header = header.Header
	block.Header.ChainID = header.ChainID
	block.Header.Hash = header.Hash
	block.Header.Protocol = header.Protocol

	block.Metadata = header.Metadata

	if level < 2 {
		block.OPG = make([]noderpc.LightOperationGroup, 0)
		return block, nil
	}
	block.OPG = header.Operations[3]

	return block, nil
}

func (r *Receiver) job(ctx context.Context, level int64) {
	block, err := r.get(ctx, level)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Error().Int64("block", level).Err(err).Msg("Receiver.get")
			r.pool.AddTask(level)
		}
		return
	}
	r.blocks <- &block
	r.inProcess.Delete(level)
}
