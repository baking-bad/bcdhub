package indexer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/migrations"
	"github.com/baking-bad/bcdhub/internal/parsers/operations"
	"github.com/baking-bad/bcdhub/internal/parsers/protocols"
	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/rollback"
	"github.com/go-pg/pg/v10"
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
		config.WithStorage(cfg.Storage, "indexer", 10, cfg.Indexer.Connections.Open, cfg.Indexer.Connections.Idle, true),
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
	}

	if err := bi.init(ctx, bi.Context.StorageDB); err != nil {
		return nil, err
	}

	return bi, nil
}

// Close -
func (bi *BlockchainIndexer) Close() error {
	close(bi.refreshTimer)
	return bi.Context.Close()
}

func (bi *BlockchainIndexer) init(ctx context.Context, db *core.Postgres) error {
	if err := NewInitializer(bi.Network, bi.Storage, bi.Blocks, db.DB, bi.RPC, bi.isPeriodic).Init(ctx); err != nil {
		return err
	}

	currentState, err := bi.Blocks.Last()
	if err != nil {
		return err
	}
	bi.state = currentState
	logger.Info().Str("network", bi.Network.String()).Msgf("Current indexer state: %d", currentState.Level)

	currentProtocol, err := bi.Protocols.Get("", currentState.Level)
	if err != nil {
		if !bi.Storage.IsRecordNotFound(err) {
			return err
		}

		header, err := bi.RPC.GetHeader(ctx, helpers.Max(1, currentState.Level))
		if err != nil {
			return err
		}

		logger.Info().Str("network", bi.Network.String()).Msgf("Creating new protocol %s starting at %d", header.Protocol, header.Level)
		currentProtocol, err = createProtocol(ctx, bi.RPC, header.ChainID, header.Protocol, header.Level)
		if err != nil {
			return err
		}

		if err := currentProtocol.Save(db.DB); err != nil {
			return err
		}
	}

	bi.currentProtocol = currentProtocol
	logger.Info().Str("network", bi.Network.String()).Msgf("Current network protocol: %s", currentProtocol.Hash)
	return nil
}

// Start -
func (bi *BlockchainIndexer) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	localSentry := helpers.GetLocalSentry()
	helpers.SetLocalTagSentry(localSentry, "network", bi.Network.String())

	wg.Add(1)
	go bi.indexBlock(ctx, wg)

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

func (bi *BlockchainIndexer) indexBlock(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case block := <-bi.receiver.Blocks():
			bi.blocks[block.Header.Level] = block

		case <-ticker.C:
			if block, ok := bi.blocks[bi.state.Level+1]; ok {
				if bi.state.Level > 0 && block.Header.Predecessor != bi.state.Hash {
					if err := bi.Rollback(ctx); err != nil {
						logger.Error().Err(err).Msg("Rollback")
					}
				} else {
					if err := bi.handleBlock(ctx, block); err != nil {
						logger.Error().Err(err).Msg("handleBlock")
					}
				}

				delete(bi.blocks, block.Header.Level)
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

	bi.indicesInit.Do(bi.createIndices)

	return nil
}

func (bi *BlockchainIndexer) handleBlock(ctx context.Context, block *Block) error {
	return bi.StorageDB.DB.RunInTransaction(ctx,
		func(tx *pg.Tx) error {
			logger.Info().Str("network", bi.Network.String()).Int64("block", block.Header.Level).Msg("indexing")

			if block.Header.Protocol != bi.currentProtocol.Hash || (bi.Network == types.Mainnet && block.Header.Level == 1) {
				logger.Info().Str("network", bi.Network.String()).Msgf("New protocol detected: %s -> %s", bi.currentProtocol.Hash, block.Header.Protocol)

				if err := bi.migrate(ctx, block.Header, tx); err != nil {
					return err
				}
			}

			store := postgres.NewStore(tx)
			if err := bi.implicitMigration(ctx, block, bi.currentProtocol, store); err != nil {
				return err
			}

			if err := bi.getDataFromBlock(block, store); err != nil {
				return err
			}

			if err := store.Save(); err != nil {
				return err
			}

			if err := bi.createBlock(block.Header, tx); err != nil {
				return err
			}
			return nil
		},
	)
}

// Rollback -
func (bi *BlockchainIndexer) Rollback(ctx context.Context) error {
	logger.Warning().Str("network", bi.Network.String()).Msgf("Rollback from %7d", bi.state.Level)

	lastLevel, err := bi.getLastRollbackBlock(ctx)
	if err != nil {
		return err
	}

	manager := rollback.NewManager(bi.RPC, bi.Storage, bi.Blocks, bi.BigMapDiffs)
	if err := manager.Rollback(ctx, bi.StorageDB.DB, bi.Network, bi.state, lastLevel); err != nil {
		return err
	}

	newState, err := bi.Blocks.Last()
	if err != nil {
		return err
	}
	bi.state = newState
	logger.Info().Str("network", bi.Network.String()).Msgf("New indexer state: %7d", bi.state.Level)
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

		block, err := bi.Blocks.Get(level - 1)
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
func (bi *BlockchainIndexer) createBlock(head noderpc.Header, tx pg.DBI) error {
	newBlock := block.Block{
		Hash:       head.Hash,
		ProtocolID: bi.currentProtocol.ID,
		Level:      head.Level,
		Timestamp:  head.Timestamp,
	}
	if err := newBlock.Save(tx); err != nil {
		return err
	}

	bi.state = newBlock
	return nil
}

func (bi *BlockchainIndexer) getDataFromBlock(block *Block, store parsers.Store) error {
	if block.Header.Level <= 1 {
		return nil
	}
	parserParams, err := operations.NewParseParams(
		bi.Context,
		operations.WithProtocol(&bi.currentProtocol),
		operations.WithHead(block.Header),
	)
	if err != nil {
		return err
	}

	for i := range block.OPG {
		parser := operations.NewGroup(parserParams)
		if err := parser.Parse(block.OPG[i], store); err != nil {
			return err
		}
	}

	return nil
}

func (bi *BlockchainIndexer) migrate(ctx context.Context, head noderpc.Header, tx pg.DBI) error {
	if bi.currentProtocol.EndLevel == 0 && head.Level > 1 {
		logger.Info().Str("network", bi.Network.String()).Msgf("Finalizing the previous protocol: %s", bi.currentProtocol.Alias)
		bi.currentProtocol.EndLevel = head.Level - 1
		if err := bi.currentProtocol.Save(bi.StorageDB.DB); err != nil {
			return err
		}
	}

	newProtocol, err := bi.Protocols.Get(head.Protocol, head.Level)
	if err != nil {
		logger.Info().Str("network", bi.Network.String()).Msgf("Creating new protocol %s starting at %d", head.Protocol, head.Level)
		newProtocol, err = createProtocol(ctx, bi.RPC, head.ChainID, head.Protocol, head.Level)
		if err != nil {
			return err
		}
		if err := newProtocol.Save(bi.StorageDB.DB); err != nil {
			return err
		}
	}

	if bi.Network == types.Mainnet && head.Level == 1 {
		if err := bi.vestingMigration(ctx, head, tx); err != nil {
			return err
		}
	} else {
		if bi.currentProtocol.SymLink == "" {
			return errors.Errorf("[%s] Protocol should be initialized", bi.Network)
		}
		if newProtocol.SymLink != bi.currentProtocol.SymLink {
			if err := bi.standartMigration(ctx, newProtocol, head, tx); err != nil {
				return err
			}
		} else {
			logger.Info().Str("network", bi.Network.String()).Msgf("Same symlink %s for %s / %s",
				newProtocol.SymLink, bi.currentProtocol.Alias, newProtocol.Alias)
		}

	}

	bi.currentProtocol = newProtocol

	bi.setUpdateTicker(0)
	logger.Info().Str("network", bi.Network.String()).Msgf("Migration to %s is completed", bi.currentProtocol.Alias)
	return nil
}

func (bi *BlockchainIndexer) implicitMigration(ctx context.Context, block *Block, protocol protocol.Protocol, store parsers.Store) error {
	if block == nil || block.Metadata == nil {
		return nil
	}

	specific, err := protocols.Get(bi.Context, protocol.Hash)
	if err != nil {
		return err
	}

	implicitParser, err := migrations.NewImplicitParser(bi.Context, bi.RPC, specific.ContractParser, protocol)
	if err != nil {
		return err
	}
	return implicitParser.Parse(ctx, *block.Metadata, block.Header, store)
}

func (bi *BlockchainIndexer) standartMigration(ctx context.Context, newProtocol protocol.Protocol, head noderpc.Header, tx pg.DBI) error {
	logger.Info().Str("network", bi.Network.String()).Msg("Try to find migrations...")

	var contracts []contract.Contract
	if err := bi.StorageDB.DB.Model((*contract.Contract)(nil)).
		Relation("Account").
		Where("tags & 4 = 0"). // except delegator contracts
		Select(&contracts); err != nil {
		return err
	}
	logger.Info().Str("network", bi.Network.String()).Msgf("Now %2d contracts are indexed", len(contracts))

	specific, err := protocols.Get(bi.Context, newProtocol.Hash)
	if err != nil {
		return err
	}

	for i := range contracts {
		if !specific.MigrationParser.IsMigratable(contracts[i].Account.Address) && newProtocol.SymLink == bcd.SymLinkJakarta {
			if _, err = bi.StorageDB.DB.Model(&contracts[i]).
				Set("jakarta_id = babylon_id").
				WherePK().
				Update(&contracts[i]); err != nil {
				return err
			}
			continue
		}

		logger.Info().Str("network", bi.Network.String()).Msgf("Migrate %s...", contracts[i].Account.Address)
		script, err := bi.RPC.GetScriptJSON(ctx, contracts[i].Account.Address, newProtocol.StartLevel)
		if err != nil {
			return err
		}

		if err := specific.MigrationParser.Parse(script, &contracts[i], bi.currentProtocol, newProtocol, head.Timestamp, tx); err != nil {
			return err
		}

		switch newProtocol.SymLink {
		case bcd.SymLinkBabylon:
			if _, err := bi.StorageDB.DB.
				Model(&contracts[i]).
				Set("babylon_id = ?babylon_id").
				WherePK().Update(&contracts[i]); err != nil {
				return err
			}
		case bcd.SymLinkJakarta:
			if _, err := bi.StorageDB.DB.
				Model(&contracts[i]).
				Set("jakarta_id = ?jakarta_id").
				WherePK().Update(&contracts[i]); err != nil {
				return err
			}
		}

	}

	// only delegator contracts
	switch newProtocol.SymLink {
	case bcd.SymLinkBabylon:
		_, err = bi.StorageDB.DB.Model((*contract.Contract)(nil)).
			Set("babylon_id = alpha_id").
			Where("tags & 4 > 0").
			Update()
	case bcd.SymLinkJakarta:
		_, err = bi.StorageDB.DB.Model((*contract.Contract)(nil)).
			Set("jakarta_id = babylon_id").
			Where("tags & 4 > 0").
			Update()
	}
	return err
}

func (bi *BlockchainIndexer) vestingMigration(ctx context.Context, head noderpc.Header, tx pg.DBI) error {
	addresses, err := bi.RPC.GetContractsByBlock(ctx, head.Level)
	if err != nil {
		return err
	}

	specific, err := protocols.Get(bi.Context, bi.currentProtocol.Hash)
	if err != nil {
		return err
	}

	p, err := migrations.NewVestingParser(bi.Context, specific.ContractParser, bi.currentProtocol)
	if err != nil {
		return err
	}

	store := postgres.NewStore(tx)

	for _, address := range addresses {
		if !bcd.IsContract(address) {
			continue
		}

		data, err := bi.RPC.GetContractData(ctx, address, head.Level)
		if err != nil {
			return err
		}

		if err := p.Parse(data, head, address, store); err != nil {
			return err
		}
	}

	return store.Save()
}

func (bi *BlockchainIndexer) reinit(ctx context.Context, cfg config.Config, indexerConfig config.IndexerConfig) error {
	bi.Context = config.NewContext(
		bi.Network,
		config.WithConfigCopy(cfg),
		config.WithStorage(cfg.Storage, "indexer", 10, cfg.Indexer.Connections.Open, cfg.Indexer.Connections.Idle, true),
		config.WithRPC(cfg.RPC),
	)
	logger.Info().Str("network", bi.Context.Network.String()).Msg("Creating indexer object...")
	bi.receiver = NewReceiver(bi.Context.RPC, 20, indexerConfig.ReceiverThreads)

	return bi.init(ctx, bi.Context.StorageDB)
}
