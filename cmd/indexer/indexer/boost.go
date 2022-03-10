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
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/rollback"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

var errBcdQuit = errors.New("bcd-quit")
var errSameLevel = errors.New("Same level")

// BoostIndexer -
type BoostIndexer struct {
	*config.Context

	rpc             noderpc.INode
	receiver        *Receiver
	state           block.Block
	currentProtocol protocol.Protocol
	blocks          map[int64]*Block

	updateTicker *time.Ticker
	Network      types.Network

	indicesInit sync.Once
}

// NewBoostIndexer -
func NewBoostIndexer(ctx context.Context, internalCtx config.Context, rpcConfig config.RPCConfig, network types.Network) (*BoostIndexer, error) {
	logger.Info().Str("network", network.String()).Msg("Creating indexer object...")

	rpcOpts := []noderpc.NodeOption{
		noderpc.WithTimeout(time.Duration(rpcConfig.Timeout) * time.Second),
	}

	if internalCtx.Config.Indexer.Cache {
		rpcOpts = append(rpcOpts, noderpc.WithCache(internalCtx.Config.SharePath, network.String()))
	}

	rpc := noderpc.NewWaitNodeRPC(
		rpcConfig.URI,
		rpcOpts...,
	)

	receiverThreadsCount := 2
	if types.Mainnet == network {
		receiverThreadsCount = 10
	}

	bi := &BoostIndexer{
		Context:  &internalCtx,
		Network:  network,
		rpc:      rpc,
		receiver: NewReceiver(rpc, 100, int64(receiverThreadsCount)),
		blocks:   make(map[int64]*Block),
	}

	if err := bi.init(ctx, bi.Context.StorageDB); err != nil {
		return nil, err
	}

	return bi, nil
}

func (bi *BoostIndexer) init(ctx context.Context, db *core.Postgres) error {
	currentState, err := bi.Blocks.Last(bi.Network)
	if err != nil {
		return err
	}
	bi.state = currentState
	logger.Info().Str("network", bi.Network.String()).Msgf("Current indexer state: %d", currentState.Level)

	currentProtocol, err := bi.Protocols.Get(bi.Network, "", currentState.Level)
	if err != nil {
		if !bi.Storage.IsRecordNotFound(err) {
			return err
		}

		header, err := bi.rpc.GetHeader(helpers.MaxInt64(1, currentState.Level))
		if err != nil {
			return err
		}
		currentProtocol, err = createProtocol(bi.rpc, bi.Network, header.Protocol, header.Level)
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

// Sync -
func (bi *BoostIndexer) Sync(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	localSentry := helpers.GetLocalSentry()
	helpers.SetLocalTagSentry(localSentry, "network", bi.Network.String())

	wg.Add(1)
	go bi.indexBlock(ctx, wg)

	bi.receiver.Start(ctx)

	// First tick
	if err := bi.process(ctx); err != nil {
		logger.Err(err)
		helpers.CatchErrorSentry(err)
	}

	everySecond := false
	duration := time.Duration(bi.currentProtocol.Constants.TimeBetweenBlocks) * time.Second
	if duration.Microseconds() <= 0 {
		duration = 10 * time.Second
	}
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
				helpers.CatchErrorSentry(err)
			}

			if everySecond {
				everySecond = false
				bi.setUpdateTicker(0)
			}
		}
	}
}

func (bi *BoostIndexer) setUpdateTicker(seconds int) {
	if bi.updateTicker != nil {
		bi.updateTicker.Stop()
	}
	var duration time.Duration
	if seconds == 0 {
		duration = time.Duration(bi.currentProtocol.Constants.TimeBetweenBlocks) * time.Second
		if duration.Microseconds() <= 0 {
			duration = 10 * time.Second
		}
	} else {
		duration = time.Duration(seconds) * time.Second
	}
	logger.Info().Str("network", bi.Network.String()).Msgf("Data will be updated every %.0f seconds", duration.Seconds())
	bi.updateTicker = time.NewTicker(duration)
}

func (bi *BoostIndexer) indexBlock(ctx context.Context, wg *sync.WaitGroup) {
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
					if !time.Now().Add(time.Duration(-5) * time.Minute).After(block.Header.Timestamp) { // Check that node is out of sync
						if err := bi.Rollback(ctx); err != nil {
							logger.Error().Err(err).Msg("Rollback")
						}
					}
				}

				if err := bi.handleBlock(ctx, block); err != nil {
					logger.Error().Err(err).Msg("handleBlock")
				}

				delete(bi.blocks, block.Header.Level)
			}
		}
	}
}

// Index -
func (bi *BoostIndexer) Index(ctx context.Context, head noderpc.Header) error {
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

func (bi *BoostIndexer) handleBlock(ctx context.Context, block *Block) error {
	result := parsers.NewResult()
	err := bi.StorageDB.DB.RunInTransaction(ctx,
		func(tx *pg.Tx) error {
			logger.Info().Str("network", bi.Network.String()).Msgf("indexing %7d block", block.Header.Level)

			if block.Header.Protocol != bi.currentProtocol.Hash || (bi.Network == types.Mainnet && block.Header.Level == 1) {
				logger.Info().Str("network", bi.Network.String()).Msgf("New protocol detected: %s -> %s", bi.currentProtocol.Hash, block.Header.Protocol)

				if err := bi.migrate(block.Header, tx); err != nil {
					return err
				}
			}

			res, err := bi.getDataFromBlock(block)
			if err != nil {
				return err
			}

			if err := res.Save(tx); err != nil {
				return err
			}

			result.Merge(res)
			if err := bi.createBlock(block.Header, tx); err != nil {
				return err
			}
			return nil
		},
	)
	return err
}

// Rollback -
func (bi *BoostIndexer) Rollback(ctx context.Context) error {
	logger.Warning().Str("network", bi.Network.String()).Msgf("Rollback from %7d", bi.state.Level)

	lastLevel, err := bi.getLastRollbackBlock()
	if err != nil {
		return err
	}

	manager := rollback.NewManager(bi.rpc, bi.Searcher, bi.Storage, bi.Blocks, bi.BigMapDiffs, bi.Transfers)
	if err := manager.Rollback(ctx, bi.StorageDB.DB, bi.state, lastLevel); err != nil {
		return err
	}

	helpers.CatchErrorSentry(errors.Errorf("[%s] Rollback from %7d to %7d", bi.Network, bi.state.Level, lastLevel))

	newState, err := bi.Blocks.Last(bi.Network)
	if err != nil {
		return err
	}
	bi.state = newState
	logger.Info().Str("network", bi.Network.String()).Msgf("New indexer state: %7d", bi.state.Level)
	logger.Info().Str("network", bi.Network.String()).Msg("Rollback finished")
	return nil
}

func (bi *BoostIndexer) getLastRollbackBlock() (int64, error) {
	var lastLevel int64
	level := bi.state.Level

	for end := false; !end; level-- {
		headAtLevel, err := bi.rpc.GetHeader(level)
		if err != nil {
			return 0, err
		}

		block, err := bi.Blocks.Get(bi.Network, level)
		if err != nil {
			return 0, err
		}

		if block.Predecessor == headAtLevel.Predecessor {
			logger.Warning().Str("network", bi.Network.String()).Msgf("Found equal predecessors at level: %7d", block.Level)
			end = true
			lastLevel = block.Level - 1
		}
	}
	return lastLevel, nil
}

func (bi *BoostIndexer) process(ctx context.Context) error {
	head, err := bi.rpc.GetHead()
	if err != nil {
		return err
	}

	if !bi.state.ValidateChainID(head.ChainID) {
		return errors.Errorf("Invalid chain_id: %s (state) != %s (head)", bi.state.ChainID, head.ChainID)
	}

	logger.Info().Str("network", bi.Network.String()).Msgf("Current node state: %7d", head.Level)
	logger.Info().Str("network", bi.Network.String()).Msgf("Current indexer state: %7d", bi.state.Level)

	if head.Level > bi.state.Level {
		if err := bi.Index(ctx, head); err != nil {
			if errors.Is(err, errBcdQuit) {
				return nil
			}
			return err
		}

		logger.Info().Str("network", bi.Network.String()).Msg("Synced")
		return nil
	} else if head.Level < bi.state.Level {
		if err := bi.Rollback(ctx); err != nil {
			return err
		}
	}

	return errSameLevel
}
func (bi *BoostIndexer) createBlock(head noderpc.Header, tx pg.DBI) error {
	newBlock := block.Block{
		Network:     bi.Network,
		Hash:        head.Hash,
		Predecessor: head.Predecessor,
		ProtocolID:  bi.currentProtocol.ID,
		ChainID:     head.ChainID,
		Level:       head.Level,
		Timestamp:   head.Timestamp,
	}
	if err := newBlock.Save(tx); err != nil {
		return err
	}

	bi.state = newBlock
	return nil
}

func (bi *BoostIndexer) getDataFromBlock(block *Block) (*parsers.Result, error) {
	result := parsers.NewResult()
	if block.Header.Level <= 1 {
		return result, nil
	}
	parserParams, err := operations.NewParseParams(
		bi.rpc,
		bi.Context,
		operations.WithProtocol(&bi.currentProtocol),
		operations.WithHead(block.Header),
		operations.WithNetwork(bi.Network),
	)
	if err != nil {
		return nil, err
	}

	for i := range block.OPG {
		parser := operations.NewGroup(parserParams)
		opgResult, err := parser.Parse(block.OPG[i])
		if err != nil {
			return nil, err
		}
		result.Merge(opgResult)
	}

	return result, nil
}

func (bi *BoostIndexer) migrate(head noderpc.Header, tx pg.DBI) error {
	if bi.currentProtocol.EndLevel == 0 && head.Level > 1 {
		logger.Info().Str("network", bi.Network.String()).Msgf("Finalizing the previous protocol: %s", bi.currentProtocol.Alias)
		bi.currentProtocol.EndLevel = head.Level - 1
		if err := bi.currentProtocol.Save(bi.StorageDB.DB); err != nil {
			return err
		}
	}

	newProtocol, err := bi.Protocols.Get(bi.Network, head.Protocol, head.Level)
	if err != nil {
		logger.Warning().Str("network", bi.Network.String()).Msgf("%s", err)
		newProtocol, err = createProtocol(bi.rpc, bi.Network, head.Protocol, head.Level)
		if err != nil {
			return err
		}
		if err := newProtocol.Save(bi.StorageDB.DB); err != nil {
			return err
		}
	}

	if bi.Network == types.Mainnet && head.Level == 1 {
		if err := bi.vestingMigration(head, tx); err != nil {
			return err
		}
	} else {
		if bi.currentProtocol.SymLink == "" {
			return errors.Errorf("[%s] Protocol should be initialized", bi.Network)
		}
		if newProtocol.SymLink != bi.currentProtocol.SymLink {
			if err := bi.standartMigration(newProtocol, head, tx); err != nil {
				return err
			}
		} else {
			logger.Info().Str("network", bi.Network.String()).Msgf("Same symlink %s for %s / %s",
				newProtocol.SymLink, bi.currentProtocol.Alias, newProtocol.Alias)
		}

		if err := bi.implicitMigration(head, tx); err != nil {
			return err
		}
	}

	bi.currentProtocol = newProtocol

	bi.setUpdateTicker(0)
	logger.Info().Str("network", bi.Network.String()).Msgf("Migration to %s is completed", bi.currentProtocol.Alias)
	return nil
}

func (bi *BoostIndexer) implicitMigration(head noderpc.Header, tx pg.DBI) error {
	metadata, err := bi.rpc.GetBlockMetadata(head.Level)
	if err != nil {
		return err
	}
	implicitParser := migrations.NewImplicitParser(bi.Context, bi.Network, bi.rpc, bi.currentProtocol)
	implicitResult, err := implicitParser.Parse(metadata, head)
	if err != nil {
		return err
	}
	if implicitResult != nil {
		return implicitResult.Save(tx)
	}
	return nil
}

func (bi *BoostIndexer) standartMigration(newProtocol protocol.Protocol, head noderpc.Header, tx pg.DBI) error {
	logger.Info().Str("network", bi.Network.String()).Msg("Try to find migrations...")
	var contracts []contract.Contract
	if err := bi.StorageDB.DB.Model((*contract.Contract)(nil)).
		Relation("Account").
		Where("contract.network = ?", bi.Network).
		Where("tags & 4 = 0"). // except delegator contracts
		Select(&contracts); err != nil {
		return err
	}
	logger.Info().Str("network", bi.Network.String()).Msgf("Now %2d contracts are indexed", len(contracts))

	migrationParser := migrations.NewMigrationParser(bi.Storage, bi.BigMapDiffs)

	for i := range contracts {
		logger.Info().Str("network", bi.Network.String()).Msgf("Migrate %s...", contracts[i].Account.Address)
		script, err := bi.rpc.GetScriptJSON(contracts[i].Account.Address, newProtocol.StartLevel)
		if err != nil {
			return err
		}

		if err := migrationParser.Parse(script, &contracts[i], bi.currentProtocol, newProtocol, head.Timestamp, tx); err != nil {
			return err
		}

		if _, err := bi.StorageDB.DB.
			Model(&contracts[i]).
			Set("alpha_id = ?alpha_id, babylon_id = ?babylon_id").
			WherePK().Update(&contracts[i]); err != nil {
			return err
		}
	}

	_, err := bi.StorageDB.DB.Model((*contract.Contract)(nil)).
		Set("babylon_id = alpha_id").
		Where("network = ?", bi.Network).
		Where("tags & 4 > 0"). // only delegator contracts
		Update()
	return err
}

func (bi *BoostIndexer) vestingMigration(head noderpc.Header, tx pg.DBI) error {
	addresses, err := bi.rpc.GetContractsByBlock(head.Level)
	if err != nil {
		return err
	}

	p := migrations.NewVestingParser(bi.Context)

	result := parsers.NewResult()

	for _, address := range addresses {
		if !bcd.IsContract(address) {
			continue
		}

		data, err := bi.rpc.GetContractData(address, head.Level)
		if err != nil {
			return err
		}

		if err := p.Parse(data, head, bi.Network, address, bi.currentProtocol, result); err != nil {
			return err
		}
	}

	return result.Save(tx)
}
