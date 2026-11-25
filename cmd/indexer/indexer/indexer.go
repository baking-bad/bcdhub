package indexer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
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
	"github.com/baking-bad/bcdhub/internal/postgres/store"
	"github.com/baking-bad/bcdhub/internal/rollback"
	"github.com/dipdup-io/workerpool"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var errBcdQuit = errors.New("bcd-quit")
var errSameLevel = errors.New("Same level")

// protocolReloadDelay is the pause between attempts to reload the current
// protocol after a rollback.
var protocolReloadDelay = time.Second

// BlockchainIndexer -
type BlockchainIndexer struct {
	*config.Context

	receiver *Receiver

	// mx serializes block processing in indexBlock with the rollback that
	// process() starts when the node is behind, and guards state,
	// currentProtocol and blocks between them.
	mx              sync.Mutex
	state           block.Block
	currentProtocol protocol.Protocol
	blocks          map[int64]*Block

	Network types.Network

	fatal        chan error
	refreshTimer chan time.Duration
	blockTime    atomic.Int64

	startLevel  int64
	isPeriodic  bool
	indicesInit sync.Once

	g   workerpool.Group
	hub *sentry.Hub
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
		config.WithStorage(ctx, cfg.Storage, "indexer", 10, time.Minute*10),
		config.WithRPC(cfg.RPC),
	)
	log.Info().Str("network", internalCtx.Network.String()).Msg("Creating indexer object...")

	hub := helpers.GetLocalSentry()
	helpers.SetLocalTagSentry(hub, "network", networkType.String())

	bi := &BlockchainIndexer{
		Context:      internalCtx,
		receiver:     NewReceiver(internalCtx.RPC, 20, indexerConfig.ReceiverThreads),
		blocks:       make(map[int64]*Block),
		Network:      networkType,
		startLevel:   indexerConfig.ResolveStartLevel(),
		isPeriodic:   indexerConfig.Periodic != nil,
		refreshTimer: make(chan time.Duration, 10),
		fatal:        make(chan error, 1),
		g:            workerpool.NewGroup(),
		hub:          hub,
	}

	if err := bi.init(ctx, bi.StorageDB); err != nil {
		return nil, err
	}

	return bi, nil
}

// Close -
func (bi *BlockchainIndexer) Close() error {
	bi.g.Wait()

	if err := bi.receiver.Close(); err != nil {
		log.Err(err).Msg("closing receiver")
	}
	if err := bi.Context.Close(); err != nil {
		log.Err(err).Msg("closing context")
	}
	return nil
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
	log.Info().Str("network", bi.Network.String()).Msgf("Current indexer state: %d", currentState.Level)

	if err := bi.refreshProtocol(ctx, currentState.Level); err != nil {
		if !bi.Storage.IsRecordNotFound(err) {
			return err
		}
		proto, err := bi.createInitialProtocol(ctx, currentState.Level)
		if err != nil {
			return err
		}
		bi.setCurrentProtocol(proto)
	}
	log.Info().Str("network", bi.Network.String()).Msgf("Current network protocol: %s", bi.currentProtocol.Hash)

	for {
		if _, err := bi.RPC.GetLevel(ctx); err == nil {
			break
		}
		log.Warn().Str("network", bi.Network.String()).Msg("waiting node rpc...")
		time.Sleep(time.Second * 15)
	}

	return nil
}

// createInitialProtocol creates the protocol of the first indexed block when
// the database has none yet.
func (bi *BlockchainIndexer) createInitialProtocol(ctx context.Context, stateLevel int64) (protocol.Protocol, error) {
	header, err := bi.RPC.GetHeader(ctx, helpers.Max(bi.startLevel, stateLevel))
	if err != nil {
		return protocol.Protocol{}, err
	}

	log.Info().
		Str("network", bi.Network.String()).
		Msgf("Creating new protocol %s starting at %d", header.Protocol, header.Level)

	currentProtocol, err := protocols.Create(ctx, bi.RPC, header)
	if err != nil {
		return protocol.Protocol{}, err
	}

	tx, err := core.NewTransaction(ctx, bi.StorageDB.DB)
	if err != nil {
		return protocol.Protocol{}, err
	}
	defer tx.Close()
	if err := tx.Protocol(ctx, &currentProtocol); err != nil {
		return protocol.Protocol{}, err
	}
	if err := tx.UpdateStats(ctx, stats.Stats{}); err != nil {
		return protocol.Protocol{}, err
	}
	if err := tx.Commit(); err != nil {
		return protocol.Protocol{}, err
	}

	if stateLevel == 0 && bi.startLevel > 1 {
		if err := bi.bootstrapImplicitContracts(ctx, header, currentProtocol); err != nil {
			return protocol.Protocol{}, err
		}
	}
	return currentProtocol, nil
}

func (bi *BlockchainIndexer) bootstrapImplicitContracts(ctx context.Context, head noderpc.Header, currentProtocol protocol.Protocol) error {
	configured := bi.Config.ImplicitContracts[bi.Network.String()]
	if len(configured) == 0 {
		return nil
	}

	specific, err := protocols.Get(bi.Context, currentProtocol.Hash)
	if err != nil {
		return err
	}

	parser := migrations.NewVestingParser(bi.Context, specific.ContractParser, currentProtocol)
	s := store.NewStore(bi.StorageDB.DB, bi.Stats)

	for _, c := range configured {
		if !bcd.IsContract(c.Address) {
			continue
		}
		data, err := bi.RPC.GetContractData(ctx, c.Address, head.Level)
		if err != nil {
			return err
		}
		if err := parser.Parse(ctx, data, head, c.Address, s); err != nil {
			return err
		}
	}

	return s.Save(ctx)
}

// Start -
func (bi *BlockchainIndexer) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer helpers.LocalCatchPanicSentry(bi.hub)
	localSentry := bi.hub

	select {
	case <-bi.fatal:
	default:
	}

	bi.g.GoCtx(ctx, bi.indexBlock)

	bi.receiver.Start(ctx)

	// First tick
	if err := bi.process(ctx); err != nil {
		if !errors.Is(err, errSameLevel) {
			bi.reportProcessError(localSentry, err)
		}
		if bi.isUnrecoverable(err) {
			return
		}
	}

	everySecond := false
	duration := bi.tickerDuration(0)
	log.Info().Str("network", bi.Network.String()).Msgf("Data will be updated every %.0f seconds", duration.Seconds())
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := bi.process(ctx); err != nil {
				if errors.Is(err, errSameLevel) {
					if !everySecond {
						everySecond = true
						bi.setUpdateTicker(5)
					}
					continue
				}
				bi.reportProcessError(localSentry, err)
				if bi.isUnrecoverable(err) {
					return
				}
			}

			if everySecond {
				everySecond = false
				bi.setUpdateTicker(0)
			}
		case d := <-bi.refreshTimer:
			// the ticker is reset here, in its owning goroutine, so that the select
			// statement picks up the new interval
			// https://go.dev/ref/spec#Select_statements
			log.Info().Str("network", bi.Network.String()).Msgf("Data will be updated every %.0f seconds", d.Seconds())
			ticker.Reset(d)

		case err := <-bi.fatal:
			log.Err(err).Str("network", bi.Network.String()).Msg("indexer stopped: unrecoverable error")
			return
		}
	}
}

// reportProcessError logs a process loop error and sends it to Sentry,
// except transient network failures: the next tick retries them anyway.
func (bi *BlockchainIndexer) reportProcessError(hub *sentry.Hub, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	if helpers.IsTransientError(err) {
		log.Warn().Err(err).Str("network", bi.Network.String()).Msg("processing: transient network error")
		return
	}
	log.Err(err).Str("network", bi.Network.String()).Msg("processing")
	helpers.LocalCatchErrorSentry(hub, err)
}

// tickerDuration returns the interval between process calls: the current protocol's
// time between blocks when seconds is 0, the given number of seconds otherwise.
func (bi *BlockchainIndexer) tickerDuration(seconds int) time.Duration {
	if seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if duration := time.Duration(bi.blockTime.Load()); duration > 0 {
		return duration
	}
	return 10 * time.Second
}

// setUpdateTicker asks the Start loop to change its interval. It is safe to call from
// any goroutine: the ticker itself is never touched outside of Start.
func (bi *BlockchainIndexer) setUpdateTicker(seconds int) {
	select {
	case bi.refreshTimer <- bi.tickerDuration(seconds):
	default:
		log.Warn().Str("network", bi.Network.String()).Msg("ticker update is dropped: refresh queue is full")
	}
}

func (bi *BlockchainIndexer) indexBlock(ctx context.Context) {
	defer helpers.LocalCatchPanicSentry(bi.hub)

	for {
		select {
		case <-ctx.Done():
			return

		case newBlock := <-bi.receiver.Blocks():
			if !bi.processBlocks(ctx, newBlock) {
				return
			}
		}
	}
}

// processBlocks stores a received block and indexes all consecutive blocks
// that are ready, rolling back when a block does not extend the current state.
// It returns false when indexing must stop.
func (bi *BlockchainIndexer) processBlocks(ctx context.Context, newBlock *Block) bool {
	bi.mx.Lock()
	defer bi.mx.Unlock()

	bi.blocks[newBlock.Header.Level] = newBlock

	nextLevel := helpers.Max(bi.state.Level+1, bi.startLevel)
	block, ok := bi.blocks[nextLevel]
	for ok {
		if bi.state.Level > 0 && block.Header.Predecessor != bi.state.Hash {
			if err := bi.rollback(ctx); err != nil {
				log.Err(err).Str("network", bi.Network.String()).Msg("rollback")
				if errors.Is(err, errDeepReorg) {
					helpers.LocalCatchErrorSentry(bi.hub, err)
					select {
					case bi.fatal <- err:
					default:
					}
					return false
				}
				// the in-memory state may not match the database after a failed
				// rollback: index nothing else until the next attempt
				delete(bi.blocks, block.Header.Level)
				return true
			}
		} else {
			if err := bi.handleBlock(ctx, block); err != nil {
				log.Err(err).
					Str("network", bi.Network.String()).
					Int64("block", block.Header.Level).
					Stack().
					Msg("handleBlock")
			}
		}

		delete(bi.blocks, block.Header.Level)
		nextLevel = helpers.Max(bi.state.Level+1, bi.startLevel)
		block, ok = bi.blocks[nextLevel]
	}
	return true
}

// Index -
func (bi *BlockchainIndexer) Index(ctx context.Context, head noderpc.Header) error {
	fromLevel := helpers.Max(bi.currentState().Level+1, bi.startLevel)
	for level := fromLevel; level <= head.Level; level++ {
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
			log.Err(err).Str("network", bi.Network.String()).Msg("can't create index")
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

	log.Info().
		Str("network", bi.Network.String()).
		Int64("processing_time_ms", time.Since(start).Milliseconds()).
		Int64("block", block.Header.Level).
		Msg("indexed")

	return nil
}

func (bi *BlockchainIndexer) parseAndSaveBlock(ctx context.Context, block *Block) error {
	store := store.NewStore(bi.StorageDB.DB, bi.Stats)
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

// setCurrentProtocol makes proto the protocol new blocks are written with and
// derives the block time from it.
func (bi *BlockchainIndexer) setCurrentProtocol(proto protocol.Protocol) {
	bi.currentProtocol = proto
	bi.blockTime.Store(proto.TimeBetweenBlocks * int64(time.Second))
}

// refreshProtocol reloads the protocol active at level from storage into memory.
// It must run after a rollback: the rollback may delete the protocol activated
// in a rolled-back block, and the cached one would then point to a removed row.
func (bi *BlockchainIndexer) refreshProtocol(ctx context.Context, level int64) error {
	proto, err := bi.Protocols.Get(ctx, "", level)
	if err != nil {
		return errors.Wrapf(err, "get protocol at level %d", level)
	}
	bi.setCurrentProtocol(proto)
	return nil
}

// reloadProtocol refreshes the current protocol after a rollback, retrying
// until it succeeds or ctx is done. Indexing cannot go on without it: the
// cached protocol may be a row the rollback has just deleted.
func (bi *BlockchainIndexer) reloadProtocol(ctx context.Context, level int64) error {
	for {
		err := bi.refreshProtocol(ctx, level)
		if err == nil {
			return nil
		}
		log.Err(err).Str("network", bi.Network.String()).Msg("reload protocol after rollback, retrying")

		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "reload protocol after rollback")
		case <-time.After(protocolReloadDelay):
		}
	}
}

// needsMigration reports whether header starts a protocol other than the current one.
func (bi *BlockchainIndexer) needsMigration(header noderpc.Header) bool {
	return header.Protocol != bi.currentProtocol.Hash || header.Level <= 1
}

func (bi *BlockchainIndexer) doMigration(ctx context.Context, header noderpc.Header) error {
	if !bi.needsMigration(header) {
		return nil
	}

	log.Info().
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
	defer tx.Close()

	migraton := protocols.NewMigration(bi.Network, bi.Context)
	newProto, err := migraton.Do(ctx, tx, bi.currentProtocol, head)
	if err != nil {
		return errors.Wrap(err, "migration.Do")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	bi.setCurrentProtocol(newProto)
	bi.setUpdateTicker(0)
	log.Info().
		Str("network", bi.Network.String()).
		Msgf("Migration to %s is completed", bi.currentProtocol.Alias)

	return nil
}

// currentState returns a copy of the indexer state safe to read from any goroutine.
func (bi *BlockchainIndexer) currentState() block.Block {
	bi.mx.Lock()
	defer bi.mx.Unlock()
	return bi.state
}

// Rollback -
func (bi *BlockchainIndexer) Rollback(ctx context.Context) error {
	bi.mx.Lock()
	defer bi.mx.Unlock()
	return bi.rollback(ctx)
}

// rollbackIfAhead rolls back if the indexer is still ahead of the node at headLevel.
// The check and the rollback run under one lock, so a rollback that indexBlock
// has just finished is not repeated.
func (bi *BlockchainIndexer) rollbackIfAhead(ctx context.Context, headLevel int64) error {
	bi.mx.Lock()
	defer bi.mx.Unlock()
	if headLevel >= bi.state.Level {
		return nil
	}
	return bi.rollback(ctx)
}

// rollback must be called with bi.mx held.
func (bi *BlockchainIndexer) rollback(ctx context.Context) error {
	log.Warn().Str("network", bi.Network.String()).Msgf("Rollback from %8d", bi.state.Level)

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
	// the rollback is committed: the state must follow it even if reloading
	// the protocol fails below
	bi.state = newState

	previousBlockTime := bi.blockTime.Load()
	if err := bi.reloadProtocol(ctx, newState.Level); err != nil {
		return err
	}
	if bi.blockTime.Load() != previousBlockTime {
		bi.setUpdateTicker(0)
	}
	log.Info().Str("network", bi.Network.String()).Msgf("New indexer state: %8d", bi.state.Level)
	log.Info().Str("network", bi.Network.String()).Msgf("Current network protocol: %s", bi.currentProtocol.Hash)
	log.Info().Str("network", bi.Network.String()).Msg("Rollback finished")
	return nil
}

var errDeepReorg = errors.New("reorg is deeper than local history")

func (bi *BlockchainIndexer) getLastRollbackBlock(ctx context.Context) (int64, error) {
	var lastLevel int64
	level := bi.state.Level

	for end := false; !end; level-- {
		if level <= bi.startLevel {
			return 0, errors.Wrapf(
				errDeepReorg,
				"reached level %d, indexing starts at %d",
				level, bi.startLevel,
			)
		}

		headAtLevel, err := bi.RPC.GetHeader(ctx, level)
		if err != nil {
			return 0, err
		}

		block, err := bi.Blocks.Get(ctx, level-1)
		if err != nil {
			return 0, err
		}

		if block.Hash == headAtLevel.Predecessor {
			log.Warn().Str("network", bi.Network.String()).Msgf("Found equal predecessors at level: %7d", block.Level)
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

	state := bi.currentState()
	if !state.Protocol.ValidateChainID(head.ChainID) {
		return errors.Errorf("Invalid chain_id: %s (state) != %s (head)", state.Protocol.ChainID, head.ChainID)
	}

	log.Info().Str("network", bi.Network.String()).Int64("node", head.Level).Int64("indexer", state.Level).Msg("current state")

	switch {
	case head.Level > state.Level:
		if err := bi.Index(ctx, head); err != nil {
			if errors.Is(err, errBcdQuit) {
				return nil
			}
			return err
		}

		log.Info().Str("network", bi.Network.String()).Msg("Synced")
		return nil
	case head.Level < state.Level:
		return bi.rollbackIfAhead(ctx, head.Level)
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
		config.WithStorage(ctx, cfg.Storage, "indexer", 10, time.Minute*10),
		config.WithRPC(cfg.RPC),
	)
	log.Info().Str("network", bi.Network.String()).Msg("Creating indexer object...")
	bi.receiver = NewReceiver(bi.RPC, 20, indexerConfig.ReceiverThreads)
	bi.startLevel = indexerConfig.ResolveStartLevel()
	bi.blocks = make(map[int64]*Block)
	return bi.init(ctx, bi.StorageDB)
}

func (bi *BlockchainIndexer) isUnrecoverable(err error) bool {
	if !errors.Is(err, errDeepReorg) {
		return false
	}
	log.Err(err).Str("network", bi.Network.String()).Msg("indexer stopped: unrecoverable error")
	return true
}
