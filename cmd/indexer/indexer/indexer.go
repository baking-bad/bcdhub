package indexer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/migrations"
	"github.com/baking-bad/bcdhub/internal/parsers/operations"
	"github.com/baking-bad/bcdhub/internal/parsers/protocols"
	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/rollback"
	"github.com/dipdup-io/workerpool"
	"github.com/pkg/errors"
)

var errBcdQuit = errors.New("bcd-quit")
var errSameLevel = errors.New("Same level")

// BlockchainIndexer -
type BlockchainIndexer struct {
	*config.Context

	receiver        *Receiver
	state           block.Block
	currentProtocol protocol.Protocol
	blocks          map[int64]*Block

	updateTicker *time.Ticker
	Network      types.Network

	refreshTimer chan struct{}

	isPeriodic  bool
	indicesInit sync.Once

	g workerpool.Group
}

// NewBlockchainIndexer -
func NewBlockchainIndexer(ctx context.Context, cfg config.Config, network string, indexerConfig config.IndexerConfig) (*BlockchainIndexer, error) {
	networkType := types.NewNetwork(network)
	if networkType == types.Empty {
		return nil, errors.Errorf("unknown network %s", network)
	}

	internalCtx := config.NewContext(
		networkType,
		config.WithConfigCopy(cfg),
		config.WithStorage(cfg.Storage, "indexer", 10),
		config.WithRPC(cfg.RPC),
	)
	logger.Info().Str("network", internalCtx.Network.String()).Msg("Creating indexer object...")

	bi := &BlockchainIndexer{
		Context:      internalCtx,
		receiver:     NewReceiver(internalCtx.RPC, 20, indexerConfig.ReceiverThreads),
		blocks:       make(map[int64]*Block),
		Network:      networkType,
		isPeriodic:   indexerConfig.Periodic != nil,
		refreshTimer: make(chan struct{}, 10),
		g:            workerpool.NewGroup(),
	}

	if err := bi.init(ctx, bi.Context.StorageDB); err != nil {
		return nil, err
	}

	return bi, nil
}

// Close -
func (bi *BlockchainIndexer) Close() error {
	bi.g.Wait()

	close(bi.refreshTimer)
	if err := bi.receiver.Close(); err != nil {
		return nil
	}
	return bi.Context.Close()
}

func (bi *BlockchainIndexer) init(ctx context.Context, db *core.Postgres) error {
	if err := NewInitializer(bi.Network, bi.Storage, bi.Blocks, db.DB, bi.RPC, bi.isPeriodic).Init(ctx); err != nil {
		return err
	}

	currentState, err := bi.Blocks.Last(ctx)
	if err != nil {
		return err
	}
	bi.state = currentState
	logger.Info().Str("network", bi.Network.String()).Msgf("Current indexer state: %d", currentState.Level)

	currentProtocol, err := bi.Protocols.Get(ctx, "", currentState.Level)
	if err != nil {
		if !bi.Storage.IsRecordNotFound(err) {
			return err
		}

		header, err := bi.RPC.GetHeader(ctx, helpers.Max(1, currentState.Level))
		if err != nil {
			return err
		}

		logger.Info().
			Str("network", bi.Network.String()).
			Msgf("Creating new protocol %s starting at %d", header.Protocol, header.Level)

		currentProtocol, err = protocols.Create(ctx, bi.RPC, header)
		if err != nil {
			return err
		}

		tx, err := core.NewTransaction(ctx, bi.StorageDB.DB)
		if err != nil {
			return err
		}
		if err := tx.Protocol(ctx, &currentProtocol); err != nil {
			return err
		}
		if err := tx.UpdateStats(ctx, stats.Stats{}); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	bi.currentProtocol = currentProtocol
	logger.Info().Str("network", bi.Network.String()).Msgf("Current network protocol: %s", currentProtocol.Hash)

	for {
		if _, err := bi.Context.RPC.GetLevel(ctx); err == nil {
			break
		}
		logger.Warning().Str("network", bi.Network.String()).Msg("waiting node rpc...")
		time.Sleep(time.Second * 15)
	}

	return nil
}

// Start -
func (bi *BlockchainIndexer) Start(ctx context.Context) {
	localSentry := helpers.GetLocalSentry()
	helpers.SetLocalTagSentry(localSentry, "network", bi.Network.String())

	bi.g.GoCtx(ctx, bi.indexBlock)

	bi.receiver.Start(ctx)

	// First tick
	if err := bi.process(ctx); err != nil {
		if !errors.Is(err, errSameLevel) {
			logger.Err(err)
			helpers.LocalCatchErrorSentry(localSentry, err)
		}
	}

	everySecond := false
	bi.setUpdateTicker(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-bi.updateTicker.C:
			if err := bi.process(ctx); err != nil {
				if errors.Is(err, errSameLevel) {
					if !everySecond {
						everySecond = true
						bi.setUpdateTicker(5)
					}
					continue
				}
				logger.Err(err)
				helpers.LocalCatchErrorSentry(localSentry, err)
			}

			if everySecond {
				everySecond = false
				bi.setUpdateTicker(0)
			}
		case <-bi.refreshTimer:
			// do nothing. refreshTimer event is for update select statement after Ticker update
			// https://go.dev/ref/spec#Select_statements
		}
	}
}

func (bi *BlockchainIndexer) setUpdateTicker(seconds int) {
	var duration time.Duration
	if seconds == 0 {
		duration = time.Duration(bi.currentProtocol.Constants.TimeBetweenBlocks) * time.Second
		if duration.Microseconds() <= 0 {
			duration = 10 * time.Second
		}
	} else {
		duration = time.Duration(seconds) * time.Second
	}
	if bi.updateTicker != nil {
		bi.updateTicker.Stop()
	}
	logger.Info().Str("network", bi.Network.String()).Msgf("Data will be updated every %.0f seconds", duration.Seconds())
	bi.updateTicker = time.NewTicker(duration)
	bi.refreshTimer <- struct{}{}
}

func (bi *BlockchainIndexer) indexBlock(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case newBlock := <-bi.receiver.Blocks():
			bi.blocks[newBlock.Header.Level] = newBlock

			block, ok := bi.blocks[bi.state.Level+1]
			for ok {
				if bi.state.Level > 0 && block.Header.Predecessor != bi.state.Hash {
					if err := bi.Rollback(ctx); err != nil {
						logger.Error().Err(err).Msg("Rollback")
					}
				} else {
					if err := bi.handleBlock(ctx, block); err != nil {
						logger.Error().Err(err).
							Str("network", bi.Network.String()).
							Int64("block", block.Header.Level).
							Stack().
							Msg("handleBlock")
					}
				}

				delete(bi.blocks, block.Header.Level)
				block, ok = bi.blocks[bi.state.Level+1]
			}
		}
	}
}

// Index -
func (bi *BlockchainIndexer) Index(ctx context.Context, head noderpc.Header) error {
	for level := bi.state.Level + 1; level <= head.Level; level++ {
		helpers.SetTagSentry("block", fmt.Sprintf("%d", level))

		select {
		case <-ctx.Done():
			return errBcdQuit
		default:
			bi.receiver.AddTask(level)
		}
	}

	bi.indicesInit.Do(func() {
		if err := bi.createIndices(ctx); err != nil {
			logger.Error().Err(err).Msg("can't create index")
		}
	})

	return nil
}

func (bi *BlockchainIndexer) handleBlock(ctx context.Context, block *Block) error {
	start := time.Now()

	if err := bi.doMigration(ctx, block.Header); err != nil {
		return errors.Wrap(err, "migration error")
	}

	if err := bi.parseAndSaveBlock(ctx, block); err != nil {
		return errors.Wrap(err, "block processing")
	}

	logger.Info().
		Str("network", bi.Network.String()).
		Int64("processing_time_ms", time.Since(start).Milliseconds()).
		Int64("block", block.Header.Level).
		Msg("indexed")

	return nil
}

func (bi *BlockchainIndexer) parseAndSaveBlock(ctx context.Context, block *Block) error {
	store := postgres.NewStore(bi.StorageDB.DB, bi.Stats)
	if err := bi.parseImplicitOperations(ctx, block, bi.currentProtocol, store); err != nil {
		return err
	}

	if err := bi.getDataFromBlock(ctx, block, store); err != nil {
		return err
	}

	if err := bi.createBlock(ctx, block.Header, store); err != nil {
		return err
	}

	if err := store.Save(ctx); err != nil {
		return err
	}

	bi.state = *store.Block
	return nil
}

func (bi *BlockchainIndexer) doMigration(ctx context.Context, header noderpc.Header) error {
	if header.Protocol == bi.currentProtocol.Hash && header.Level > 1 {
		return nil
	}

	logger.Info().
		Str("network", bi.Network.String()).
		Int64("block", header.Level).
		Msgf("New protocol detected: %s -> %s", bi.currentProtocol.Hash, header.Protocol)

	return bi.migrate(ctx, header)
}

func (bi *BlockchainIndexer) migrate(ctx context.Context, head noderpc.Header) error {
	tx, err := core.NewTransaction(ctx, bi.StorageDB.DB)
	if err != nil {
		return errors.Wrap(err, "create postgres transaction")
	}

	migraton := protocols.NewMigration(bi.Network, bi.Context)
	newProto, err := migraton.Do(ctx, tx, bi.currentProtocol, head)
	if err != nil {
		return errors.Wrap(err, "migration.Do")
	}

	bi.currentProtocol = newProto

	if err := tx.Commit(); err != nil {
		return err
	}

	bi.setUpdateTicker(0)
	logger.Info().
		Str("network", bi.Network.String()).
		Msgf("Migration to %s is completed", bi.currentProtocol.Alias)

	return nil
}

// Rollback -
func (bi *BlockchainIndexer) Rollback(ctx context.Context) error {
	logger.Warning().Str("network", bi.Network.String()).Msgf("Rollback from %8d", bi.state.Level)

	lastLevel, err := bi.getLastRollbackBlock(ctx)
	if err != nil {
		return err
	}

	saver, err := postgres.NewRollback(bi.StorageDB.DB)
	if err != nil {
		return err
	}
	manager := rollback.NewManager(bi.Storage, bi.Blocks, saver, bi.Stats)
	if err := manager.Rollback(ctx, bi.Network, bi.state, lastLevel); err != nil {
		return err
	}

	newState, err := bi.Blocks.Last(ctx)
	if err != nil {
		return err
	}
	bi.state = newState
	logger.Info().Str("network", bi.Network.String()).Msgf("New indexer state: %8d", bi.state.Level)
	logger.Info().Str("network", bi.Network.String()).Msg("Rollback finished")
	return nil
}

func (bi *BlockchainIndexer) getLastRollbackBlock(ctx context.Context) (int64, error) {
	var lastLevel int64
	level := bi.state.Level

	for end := false; !end; level-- {
		headAtLevel, err := bi.RPC.GetHeader(ctx, level)
		if err != nil {
			return 0, err
		}

		block, err := bi.Blocks.Get(ctx, level-1)
		if err != nil {
			return 0, err
		}

		if block.Hash == headAtLevel.Predecessor {
			logger.Warning().Str("network", bi.Network.String()).Msgf("Found equal predecessors at level: %7d", block.Level)
			end = true
			lastLevel = block.Level
		}
	}
	return lastLevel, nil
}

func (bi *BlockchainIndexer) process(ctx context.Context) error {
	head, err := bi.RPC.GetHead(ctx)
	if err != nil {
		return err
	}

	if !bi.state.Protocol.ValidateChainID(head.ChainID) {
		return errors.Errorf("Invalid chain_id: %s (state) != %s (head)", bi.state.Protocol.ChainID, head.ChainID)
	}

	logger.Info().Str("network", bi.Network.String()).Int64("node", head.Level).Int64("indexer", bi.state.Level).Msg("current state")

	switch {
	case head.Level > bi.state.Level:
		if err := bi.Index(ctx, head); err != nil {
			if errors.Is(err, errBcdQuit) {
				return nil
			}
			return err
		}

		logger.Info().Str("network", bi.Network.String()).Msg("Synced")
		return nil
	case head.Level < bi.state.Level:
		return bi.Rollback(ctx)
	default:
		return errSameLevel
	}
}
func (bi *BlockchainIndexer) createBlock(ctx context.Context, head noderpc.Header, store parsers.Store) error {
	newBlock := block.Block{
		Hash:       head.Hash,
		ProtocolID: bi.currentProtocol.ID,
		Level:      head.Level,
		Timestamp:  head.Timestamp,
	}
	store.SetBlock(&newBlock)
	return nil
}

func (bi *BlockchainIndexer) getDataFromBlock(ctx context.Context, block *Block, store parsers.Store) error {
	if block.Header.Level <= 1 {
		return nil
	}
	parserParams, err := operations.NewParseParams(
		ctx,
		bi.Context,
		operations.WithProtocol(&bi.currentProtocol),
		operations.WithHead(block.Header),
	)
	if err != nil {
		return err
	}

	for i := range block.OPG {
		parser := operations.NewGroup(parserParams)
		if err := parser.Parse(ctx, block.OPG[i], store); err != nil {
			return err
		}
	}

	return nil
}
func (bi *BlockchainIndexer) parseImplicitOperations(ctx context.Context, block *Block, protocol protocol.Protocol, store parsers.Store) error {
	if block == nil || block.Metadata == nil {
		return nil
	}

	specific, err := protocols.Get(bi.Context, protocol.Hash)
	if err != nil {
		return err
	}

	implicitParser, err := migrations.NewImplicitParser(bi.Context, bi.RPC, specific.ContractParser, protocol, bi.Contracts)
	if err != nil {
		return err
	}
	return implicitParser.Parse(ctx, *block.Metadata, block.Header, store)
}

func (bi *BlockchainIndexer) reinit(ctx context.Context, cfg config.Config, indexerConfig config.IndexerConfig) error {
	bi.Context = config.NewContext(
		bi.Network,
		config.WithConfigCopy(cfg),
		config.WithStorage(cfg.Storage, "indexer", 10),
		config.WithRPC(cfg.RPC),
	)
	logger.Info().Str("network", bi.Context.Network.String()).Msg("Creating indexer object...")
	bi.receiver = NewReceiver(bi.Context.RPC, 20, indexerConfig.ReceiverThreads)

	bi.refreshTimer = make(chan struct{}, 10)
	return bi.init(ctx, bi.Context.StorageDB)
}
