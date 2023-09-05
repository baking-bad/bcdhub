package indexer

import (
	"context"
	"errors"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/periodic"
	"github.com/dipdup-io/workerpool"
)

// PeriodicIndexer -
type PeriodicIndexer struct {
	indexer       *BlockchainIndexer
	indexerCancel context.CancelFunc

	cfg        config.Config
	indexerCfg config.IndexerConfig

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
		time.Sleep(time.Second)
	}

	setUrlToConfig(&p.cfg, worker.URL(), network)

	bi, err := NewBlockchainIndexer(ctx, cfg, network, indexerCfg)
	if err != nil {
		return nil, err
	}
	p.indexer = bi

	return p, nil
}

// Start -
func (p *PeriodicIndexer) Start(ctx context.Context) {
	indexerCtx, indexerCancel := context.WithCancel(ctx)
	p.indexerCancel = indexerCancel
	p.indexer.Start(indexerCtx)
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
	logger.Warning().Str("network", network).Str("url", url).Msg("cancelling indexer due to URL changing...")
	p.indexerCancel()

	if err := p.indexer.Close(); err != nil {
		return err
	}

	setUrlToConfig(&p.cfg, url, network)

	if err := p.indexer.reinit(ctx, p.cfg, p.indexerCfg); err != nil {
		return err
	}

	indexerCtx, indexerCancel := context.WithCancel(ctx)
	p.indexerCancel = indexerCancel
	p.g.GoCtx(indexerCtx, p.indexer.Start)

	return nil
}

func setUrlToConfig(cfg *config.Config, url string, network string) {
	if value, ok := cfg.RPC[network]; ok {
		value.URI = url
		cfg.RPC[network] = value
	}
}
