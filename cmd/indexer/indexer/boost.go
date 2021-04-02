package indexer

import (
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/index"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/operations"
	pgBigMapAction "github.com/baking-bad/bcdhub/internal/postgres/bigmapaction"
	pgBigMapDiff "github.com/baking-bad/bcdhub/internal/postgres/bigmapdiff"
	pgBlock "github.com/baking-bad/bcdhub/internal/postgres/block"
	pgContract "github.com/baking-bad/bcdhub/internal/postgres/contract"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	pgMigration "github.com/baking-bad/bcdhub/internal/postgres/migration"
	pgOperation "github.com/baking-bad/bcdhub/internal/postgres/operation"
	pgProtocol "github.com/baking-bad/bcdhub/internal/postgres/protocol"
	pgTezosDomain "github.com/baking-bad/bcdhub/internal/postgres/tezosdomain"
	pgTokenBalance "github.com/baking-bad/bcdhub/internal/postgres/tokenbalance"
	pgTransfer "github.com/baking-bad/bcdhub/internal/postgres/transfer"
	pgTZIP "github.com/baking-bad/bcdhub/internal/postgres/tzip"
	"github.com/baking-bad/bcdhub/internal/rollback"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var errBcdQuit = errors.New("bcd-quit")
var errRollback = errors.New("rollback")
var errSameLevel = errors.New("Same level")

// BoostIndexer -
type BoostIndexer struct {
	Searcher      search.Searcher
	Storage       models.GeneralRepository
	BigMapActions bigmapaction.Repository
	BigMapDiffs   bigmapdiff.Repository
	Blocks        block.Repository
	Contracts     contract.Repository
	Migrations    migration.Repository
	Operations    operation.Repository
	Protocols     protocol.Repository
	TezosDomains  tezosdomain.Repository
	TokenBalances tokenbalance.Repository
	Transfers     transfer.Repository
	TZIP          tzip.Repository

	rpc             noderpc.INode
	externalIndexer index.Indexer
	messageQueue    mq.Mediator
	state           block.Block
	currentProtocol protocol.Protocol
	cfg             config.Config
	pg              *core.Postgres

	updateTicker        *time.Ticker
	stop                chan struct{}
	Network             string
	boost               bool
	skipDelegatorBlocks bool
	stopped             bool
}

func (bi *BoostIndexer) fetchExternalProtocols() error {
	logger.WithNetwork(bi.Network).Info("Fetching external protocols")
	existingProtocols, err := bi.Protocols.GetByNetworkWithSort(bi.Network, "start_level", "desc")
	if err != nil {
		return err
	}

	exists := make(map[string]bool)
	for _, existingProtocol := range existingProtocols {
		exists[existingProtocol.Hash] = true
	}

	extProtocols, err := bi.externalIndexer.GetProtocols()
	if err != nil {
		return err
	}

	protocols := make([]models.Model, 0)
	for i := range extProtocols {
		if _, ok := exists[extProtocols[i].Hash]; ok {
			continue
		}
		symLink, err := bcd.GetProtoSymLink(extProtocols[i].Hash)
		if err != nil {
			return err
		}
		alias := extProtocols[i].Alias
		if alias == "" {
			alias = extProtocols[i].Hash[:8]
		}

		newProtocol := &protocol.Protocol{
			Hash:       extProtocols[i].Hash,
			Alias:      alias,
			StartLevel: extProtocols[i].StartLevel,
			EndLevel:   extProtocols[i].LastLevel,
			SymLink:    symLink,
			Network:    bi.Network,
			Constants: &protocol.Constants{
				CostPerByte:                  extProtocols[i].Constants.CostPerByte,
				HardStorageLimitPerOperation: extProtocols[i].Constants.HardStorageLimitPerOperation,
				HardGasLimitPerOperation:     extProtocols[i].Constants.HardGasLimitPerOperation,
				TimeBetweenBlocks:            extProtocols[i].Constants.TimeBetweenBlocks,
			},
		}

		protocols = append(protocols, newProtocol)
		logger.WithNetwork(bi.Network).Infof("Fetched %s", alias)
	}

	return bi.Storage.Save(protocols)
}

// NewBoostIndexer -
func NewBoostIndexer(cfg config.Config, network string, opts ...BoostIndexerOption) (*BoostIndexer, error) {
	logger.WithNetwork(network).Info("Creating indexer object...")
	pg := core.WaitNew(cfg.Storage.Postgres, "indexer", 10)

	rpcProvider, ok := cfg.RPC[network]
	if !ok {
		pg.Close()
		return nil, errors.Errorf("Unknown network %s", network)
	}
	rpc := noderpc.NewWaitNodeRPC(
		rpcProvider.URI,
		noderpc.WithTimeout(time.Duration(rpcProvider.Timeout)*time.Second),
	)

	bi := &BoostIndexer{
		Searcher:      elastic.WaitNew(cfg.Storage.Elastic, 10),
		Storage:       pg,
		BigMapActions: pgBigMapAction.NewStorage(pg),
		BigMapDiffs:   pgBigMapDiff.NewStorage(pg),
		Blocks:        pgBlock.NewStorage(pg),
		Contracts:     pgContract.NewStorage(pg),
		Migrations:    pgMigration.NewStorage(pg),
		Operations:    pgOperation.NewStorage(pg),
		Protocols:     pgProtocol.NewStorage(pg),
		TezosDomains:  pgTezosDomain.NewStorage(pg),
		TokenBalances: pgTokenBalance.NewStorage(pg),
		Transfers:     pgTransfer.NewStorage(pg),
		TZIP:          pgTZIP.NewStorage(pg),
		Network:       network,
		rpc:           rpc,
		messageQueue:  mq.New(cfg.RabbitMQ.URI, cfg.Indexer.ProjectName, cfg.Indexer.MQ.NeedPublisher, 10),
		stop:          make(chan struct{}),
		cfg:           cfg,
		pg:            pg,
	}

	for _, opt := range opts {
		opt(bi)
	}

	if err := bi.init(pg); err != nil {
		bi.messageQueue.Close()
		pg.Close()
		return nil, err
	}

	return bi, nil
}

func addTriggers(db *core.Postgres) error {
	files, err := ioutil.ReadDir("triggers")
	if err != nil {
		return err
	}
	for i := range files {
		path := fmt.Sprintf("triggers/%s", files[i].Name())
		raw, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if err := db.AddTrigger(string(raw)); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) init(db *core.Postgres) error {
	if err := bi.Storage.CreateIndexes(); err != nil {
		return err
	}

	if err := addTriggers(db); err != nil {
		return err
	}

	if bi.boost {
		if err := bi.fetchExternalProtocols(); err != nil {
			return err
		}
	}

	currentState, err := bi.Blocks.Last(bi.Network)
	if err != nil {
		return err
	}
	bi.state = currentState
	logger.WithNetwork(bi.Network).Infof("Current indexer state: %d", currentState.Level)

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

		if err := bi.Storage.Save([]models.Model{&currentProtocol}); err != nil {
			return err
		}
	}

	bi.currentProtocol = currentProtocol
	logger.WithNetwork(bi.Network).Infof("Current network protocol: %s", currentProtocol.Hash)
	return nil
}

// Sync -
func (bi *BoostIndexer) Sync(wg *sync.WaitGroup) {
	defer wg.Done()

	bi.stopped = false
	localSentry := helpers.GetLocalSentry()
	helpers.SetLocalTagSentry(localSentry, "network", bi.Network)

	// First tick
	if err := bi.process(); err != nil {
		logger.Error(err)
		helpers.CatchErrorSentry(err)
	}
	if bi.stopped {
		return
	}

	everySecond := false
	duration := time.Duration(bi.currentProtocol.Constants.TimeBetweenBlocks) * time.Second
	if duration.Microseconds() <= 0 {
		duration = 10 * time.Second
	}
	bi.setUpdateTicker(0)
	for {
		select {
		case <-bi.stop:
			bi.stopped = true
			bi.messageQueue.Close()
			bi.Storage.(*core.Postgres).Close()
			return
		case <-bi.updateTicker.C:
			if err := bi.process(); err != nil {
				if errors.Is(err, errSameLevel) {
					if !everySecond {
						everySecond = true
						bi.setUpdateTicker(5)
					}
					continue
				}
				logger.Error(err)
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
	logger.WithNetwork(bi.Network).Infof("Data will be updated every %.0f seconds", duration.Seconds())
	bi.updateTicker = time.NewTicker(duration)
}

// Stop -
func (bi *BoostIndexer) Stop() {
	bi.stop <- struct{}{}
}

// Index -
func (bi *BoostIndexer) Index(levels []int64) error {
	if len(levels) == 0 {
		return nil
	}
	helpers.SetTagSentry("network", bi.Network)

	for _, level := range levels {
		helpers.SetTagSentry("block", fmt.Sprintf("%d", level))

		select {
		case <-bi.stop:
			bi.stopped = true
			bi.messageQueue.Close()
			bi.Storage.(*core.Postgres).Close()
			return errBcdQuit
		default:
		}

		head, err := bi.rpc.GetHeader(level)
		if err != nil {
			return err
		}

		if bi.state.Level > 0 && head.Predecessor != bi.state.Hash && !bi.boost {
			return errRollback
		}

		if err := bi.handleBlock(head); err != nil {
			return err
		}

	}
	return nil
}

func (bi *BoostIndexer) handleBlock(head noderpc.Header) error {
	parsed := make([]models.Model, 0)
	err := bi.pg.DB.Transaction(
		func(tx *gorm.DB) error {
			logger.WithNetwork(bi.Network).Infof("indexing %d block", head.Level)

			if head.Protocol != bi.currentProtocol.Hash {
				logger.WithNetwork(bi.Network).Infof("New protocol detected: %s -> %s", bi.currentProtocol.Hash, head.Protocol)

				migrations, err := bi.migrate(head, tx)
				if err != nil {
					return err
				}
				parsed = append(parsed, migrations...)
			}

			res, err := bi.getDataFromBlock(bi.Network, head)
			if err != nil {
				return err
			}

			for i := range res {
				if err := res[i].Save(tx); err != nil {
					return err
				}
			}

			parsed = append(parsed, res...)
			if err := bi.createBlock(head, tx); err != nil {
				return err
			}
			return nil
		},
	)

	if err != nil {
		return err
	}

	return bi.sendToQueue(parsed)
}

// Rollback -
func (bi *BoostIndexer) Rollback() error {
	logger.WithNetwork(bi.Network).Warningf("Rollback from %d", bi.state.Level)

	lastLevel, err := bi.getLastRollbackBlock()
	if err != nil {
		return err
	}

	manager := rollback.NewManager(bi.Searcher, bi.Storage, bi.BigMapDiffs, bi.Transfers)
	if err := manager.Rollback(bi.pg.DB, bi.state, lastLevel); err != nil {
		return err
	}

	helpers.CatchErrorSentry(errors.Errorf("[%s] Rollback from %d to %d", bi.Network, bi.state.Level, lastLevel))

	newState, err := bi.Blocks.Last(bi.Network)
	if err != nil {
		return err
	}
	bi.state = newState
	logger.WithNetwork(bi.Network).Infof("New indexer state: %d", bi.state.Level)
	logger.WithNetwork(bi.Network).Info("Rollback finished")
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
			logger.WithNetwork(bi.Network).Warnf("Found equal predecessors at level: %d", block.Level)
			end = true
			lastLevel = block.Level - 1
		}
	}
	return lastLevel, nil
}

func (bi *BoostIndexer) getBoostBlocks(head noderpc.Header) ([]int64, error) {
	levels, err := bi.externalIndexer.GetContractOperationBlocks(bi.state.Level, head.Level, bi.skipDelegatorBlocks)
	if err != nil {
		return nil, err
	}

	protocols, err := bi.externalIndexer.GetProtocols()
	if err != nil {
		return nil, err
	}

	protocolLevels := make([]int64, 0)
	for i := range protocols {
		if protocols[i].StartLevel > bi.state.Level && protocols[i].StartLevel > 0 {
			protocolLevels = append(protocolLevels, protocols[i].StartLevel)
		}
	}

	result := helpers.Merge2ArraysInt64(levels, protocolLevels)
	return result, err
}

func (bi *BoostIndexer) process() error {
	head, err := bi.rpc.GetHead()
	if err != nil {
		return err
	}

	if !bi.state.ValidateChainID(head.ChainID) {
		return errors.Errorf("Invalid chain_id: %s (state) != %s (head)", bi.state.ChainID, head.ChainID)
	}

	logger.WithNetwork(bi.Network).Infof("Current node state: %d", head.Level)
	logger.WithNetwork(bi.Network).Infof("Current indexer state: %d", bi.state.Level)

	if head.Level > bi.state.Level {
		levels := make([]int64, 0)
		if bi.boost {
			levels, err = bi.getBoostBlocks(head)
			if err != nil {
				return err
			}
		} else {
			for i := bi.state.Level + 1; i <= head.Level; i++ {
				levels = append(levels, i)
			}
		}

		logger.WithNetwork(bi.Network).Infof("Found %d new levels", len(levels))

		if err := bi.Index(levels); err != nil {
			if errors.Is(err, errBcdQuit) {
				return nil
			}
			if errors.Is(err, errRollback) {
				if !time.Now().Add(time.Duration(-5) * time.Minute).After(head.Timestamp) { // Check that node is out of sync
					if err := bi.Rollback(); err != nil {
						return err
					}
				}
				return nil
			}
			return err
		}

		if bi.boost {
			bi.boost = false
		}
		logger.WithNetwork(bi.Network).Info("Synced")
		return nil
	} else if head.Level < bi.state.Level {
		if err := bi.Rollback(); err != nil {
			return err
		}
	}

	return errSameLevel
}

func (bi *BoostIndexer) createBlock(head noderpc.Header, tx *gorm.DB) error {
	newBlock := block.Block{
		Network:     bi.Network,
		Hash:        head.Hash,
		Predecessor: head.Predecessor,
		Protocol:    head.Protocol,
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

func (bi *BoostIndexer) sendToQueue(items []models.Model) error {
	for i := range items {
		if err := bi.messageQueue.Send(items[i]); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) getDataFromBlock(network string, head noderpc.Header) ([]models.Model, error) {
	if head.Level <= 1 {
		return nil, nil
	}
	opg, err := bi.rpc.GetOPG(head.Level)
	if err != nil {
		return nil, err
	}

	result := make([]models.Model, 0)
	for i := range opg {
		parser := operations.NewGroup(operations.NewParseParams(
			bi.rpc,
			bi.Storage, bi.Contracts, bi.BigMapDiffs, bi.Blocks, bi.TZIP, bi.TokenBalances,
			operations.WithConstants(*bi.currentProtocol.Constants),
			operations.WithHead(head),
			operations.WithIPFSGateways(bi.cfg.IPFSGateways),
			operations.WithShareDirectory(bi.cfg.SharePath),
			operations.WithNetwork(network),
		))
		parsed, err := parser.Parse(opg[i])
		if err != nil {
			return nil, err
		}
		result = append(result, parsed...)
	}

	return result, nil
}

func (bi *BoostIndexer) migrate(head noderpc.Header, tx *gorm.DB) ([]models.Model, error) {
	if bi.currentProtocol.EndLevel == 0 && head.Level > 1 {
		logger.WithNetwork(bi.Network).Infof("Finalizing the previous protocol: %s", bi.currentProtocol.Alias)
		bi.currentProtocol.EndLevel = head.Level - 1
		if err := bi.currentProtocol.Save(tx); err != nil {
			return nil, err
		}
	}

	newProtocol, err := bi.Protocols.Get(bi.Network, head.Protocol, head.Level)
	if err != nil {
		logger.Warning("%s", err)
		newProtocol, err = createProtocol(bi.rpc, bi.Network, head.Protocol, head.Level)
		if err != nil {
			return nil, err
		}
		if err := newProtocol.Save(tx); err != nil {
			return nil, err
		}
	}

	result := make([]models.Model, 0)
	if bi.Network == consts.Mainnet && head.Level == 1 {
		items, err := bi.vestingMigration(head, tx)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	} else {
		if bi.currentProtocol.SymLink == "" {
			return nil, errors.Errorf("[%s] Protocol should be initialized", bi.Network)
		}
		if newProtocol.SymLink != bi.currentProtocol.SymLink {
			if err := bi.standartMigration(newProtocol, head, tx); err != nil {
				return nil, err
			}
		} else {
			logger.WithNetwork(bi.Network).Infof("Same symlink %s for %s / %s",
				newProtocol.SymLink, bi.currentProtocol.Alias, newProtocol.Alias)
		}
	}

	bi.currentProtocol = newProtocol

	bi.setUpdateTicker(0)
	logger.WithNetwork(bi.Network).Infof("Migration to %s is completed", bi.currentProtocol.Alias)
	return result, nil
}

func (bi *BoostIndexer) standartMigration(newProtocol protocol.Protocol, head noderpc.Header, tx *gorm.DB) error {
	logger.WithNetwork(bi.Network).Info("Try to find migrations...")
	contracts, err := bi.Contracts.GetMany(map[string]interface{}{
		"network": bi.Network,
	})
	if err != nil {
		return err
	}
	logger.WithNetwork(bi.Network).Infof("Now %d contracts are indexed", len(contracts))

	p := parsers.NewMigrationParser(bi.Storage, bi.BigMapDiffs, bi.cfg.SharePath)

	for i := range contracts {
		logger.WithNetwork(bi.Network).Infof("Migrate %s...", contracts[i].Address)
		script, err := bi.rpc.GetScriptJSON(contracts[i].Address, newProtocol.StartLevel)
		if err != nil {
			return err
		}

		if err := p.Parse(script, contracts[i], bi.currentProtocol, newProtocol, head.Timestamp, tx); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) vestingMigration(head noderpc.Header, tx *gorm.DB) ([]models.Model, error) {
	addresses, err := bi.rpc.GetContractsByBlock(head.Level)
	if err != nil {
		return nil, err
	}

	p := parsers.NewVestingParser(bi.cfg.SharePath)

	items := make([]models.Model, 0)
	for _, address := range addresses {
		if !bcd.IsContract(address) {
			continue
		}

		data, err := bi.rpc.GetContractData(address, head.Level)
		if err != nil {
			return nil, err
		}

		parsed, err := p.Parse(data, head, bi.Network, address)
		if err != nil {
			return nil, err
		}
		for i := range parsed {
			if err := parsed[i].Save(tx); err != nil {
				return nil, err
			}
		}
		items = append(items, parsed...)
	}

	return items, nil
}
