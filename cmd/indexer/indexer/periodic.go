package indexer

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/periodic"
	"github.com/dipdup-io/workerpool"
	"github.com/rs/zerolog/log"
)

// PeriodicIndexer -
type PeriodicIndexer struct {
	indexer       *BlockchainIndexer
	indexerCancel context.CancelFunc

	cfg         config.Config
	mx          sync.Mutex
	indexerCfg  config.IndexerConfig
	indexerDone chan struct{}

	worker *periodic.Worker
	g      workerpool.Group
}

// NewPeriodicIndexer -
func NewPeriodicIndexer(
	ctx context.Context,
	network string,
	cfg config.Config,
	indexerCfg config.IndexerConfig,
	g workerpool.Group,
) (*PeriodicIndexer, error) {
	if indexerCfg.Periodic == nil {
		return nil, errors.New("not periodic")
	}

	p := &PeriodicIndexer{
		cfg:        cfg,
		indexerCfg: indexerCfg,
		g:          g,
	}

	worker, err := periodic.New(*indexerCfg.Periodic, types.NewNetwork(network), p.handleUrlChanged)
	if err != nil {
		return nil, err
	}
	p.worker = worker
	p.worker.Start(ctx)

	for worker.URL() == "" {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(time.Second)
		}
	}

	setUrlToConfig(&p.cfg, worker.URL(), network)

	bi, err := NewBlockchainIndexer(ctx, p.cfg, network, indexerCfg)
	if err != nil {
		return nil, err
	}
	p.indexer = bi

	return p, nil
}

// Start -
func (p *PeriodicIndexer) Start(ctx context.Context) {
	p.runIndexer(ctx)
}

// runIndexer spawns the indexer loop in the worker group and records its cancel func
// and exit channel, so that handleUrlChanged can stop it and wait for it to finish.
func (p *PeriodicIndexer) runIndexer(ctx context.Context) {
	done := make(chan struct{})
	indexerCtx, indexerCancel := context.WithCancel(ctx)

	p.mx.Lock()
	p.indexerDone = done
	p.indexerCancel = indexerCancel
	p.mx.Unlock()

	p.g.GoCtx(indexerCtx, func(ctx context.Context) {
		defer close(done)
		p.indexer.Start(ctx)
	})
}

// Close -
func (p *PeriodicIndexer) Close() error {
	if err := p.worker.Close(); err != nil {
		return err
	}
	return p.indexer.Close()
}

// Index -
func (p *PeriodicIndexer) Index(ctx context.Context, head noderpc.Header) error {
	return p.indexer.Index(ctx, head)
}

// Rollback -
func (p *PeriodicIndexer) Rollback(ctx context.Context) error {
	return p.indexer.Rollback(ctx)
}

func (p *PeriodicIndexer) handleUrlChanged(ctx context.Context, network, url string) error {
	log.Warn().Str("network", network).Str("url", url).Msg("cancelling indexer due to URL changing...")
	if p.indexer == nil {
		return errors.New("indexer is nil")
	}
	if err := p.stopIndexer(ctx); err != nil {
		return err
	}

	setUrlToConfig(&p.cfg, url, network)

	if err := p.indexer.reinit(ctx, p.cfg, p.indexerCfg); err != nil {
		return err
	}

	p.runIndexer(ctx)
	return nil
}

func setUrlToConfig(cfg *config.Config, url string, network string) {
	if value, ok := cfg.RPC[network]; ok {
		value.URI = url
		cfg.RPC[network] = value
	}
}

// stopIndexer cancels the running indexer loop and waits for its exit.
func (p *PeriodicIndexer) stopIndexer(ctx context.Context) error {
	p.mx.Lock()
	cancel, done := p.indexerCancel, p.indexerDone
	p.mx.Unlock()

	if cancel == nil {
		return errors.New("indexer is not running")
	}
	cancel()

	select {
	case <-done:
		return p.indexer.Close()
	case <-ctx.Done():
		return ctx.Err()
	}
}
