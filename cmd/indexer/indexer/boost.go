package indexer

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/index"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/rollback"
	"github.com/pkg/errors"
)

var errBcdQuit = errors.New("bcd-quit")
var errRollback = errors.New("rollback")

// BoostIndexer -
type BoostIndexer struct {
	Network string

	rpc                 noderpc.INode
	es                  elastic.IElastic
	externalIndexer     index.Indexer
	state               models.Block
	currentProtocol     models.Protocol
	messageQueue        *mq.QueueManager
	filesDirectory      string
	boost               bool
	interfaces          map[string]kinds.ContractKind
	tokenViews          parsers.TokenViews
	skipDelegatorBlocks bool

	stop    chan struct{}
	stopped bool
}

func (bi *BoostIndexer) fetchExternalProtocols() error {
	logger.Info("[%s] Fetching external protocols", bi.Network)
	var existingProtocols []models.Protocol
	if err := bi.es.GetByNetworkWithSort(bi.Network, "start_level", "desc", &existingProtocols); err != nil {
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

	protocols := make([]elastic.Model, 0)
	for i := range extProtocols {
		if _, ok := exists[extProtocols[i].Hash]; ok {
			continue
		}
		symLink, err := meta.GetProtoSymLink(extProtocols[i].Hash)
		if err != nil {
			return err
		}
		alias := extProtocols[i].Alias
		if alias == "" {
			alias = extProtocols[i].Hash[:8]
		}

		newProtocol := &models.Protocol{
			ID:         helpers.GenerateID(),
			Hash:       extProtocols[i].Hash,
			Alias:      alias,
			StartLevel: extProtocols[i].StartLevel,
			EndLevel:   extProtocols[i].LastLevel,
			SymLink:    symLink,
			Network:    bi.Network,
		}

		protocolConstants := models.Constants{}
		if newProtocol.StartLevel != newProtocol.EndLevel || newProtocol.EndLevel != 0 {
			constants, err := bi.rpc.GetNetworkConstants(extProtocols[i].StartLevel)
			if err != nil {
				return err
			}
			protocolConstants.CostPerByte = constants.Get("cost_per_byte").Int()
			protocolConstants.HardGasLimitPerOperation = constants.Get("hard_gas_limit_per_operation").Int()
			protocolConstants.HardStorageLimitPerOperation = constants.Get("hard_storage_limit_per_operation").Int()
			protocolConstants.TimeBetweenBlocks = constants.Get("time_between_blocks.0").Int()
		}
		newProtocol.Constants = protocolConstants

		protocols = append(protocols, newProtocol)
		logger.Info("[%s] Fetched %s", bi.Network, alias)
	}

	return bi.es.BulkInsert(protocols)
}

// NewBoostIndexer -
func NewBoostIndexer(cfg config.Config, network string, opts ...BoostIndexerOption) (*BoostIndexer, error) {
	logger.Info("[%s] Creating indexer object...", network)
	es := elastic.WaitNew([]string{cfg.Elastic.URI}, cfg.Elastic.Timeout)
	rpcProvider, ok := cfg.RPC[network]
	if !ok {
		return nil, errors.Errorf("Unknown network %s", network)
	}
	rpc := noderpc.NewWaitNodeRPC(
		rpcProvider.URI,
		noderpc.WithTimeout(time.Duration(rpcProvider.Timeout)*time.Second),
	)

	db, err := database.New(cfg.DB.ConnString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	messageQueue, err := mq.NewQueueManager(cfg.RabbitMQ.URI, cfg.Indexer.ProjectName, cfg.RabbitMQ.NeedPublisher)
	if err != nil {
		return nil, err
	}

	interfaces, err := kinds.Load()
	if err != nil {
		return nil, err
	}

	tokenViews, err := parsers.NewTokenViews(db)
	if err != nil {
		return nil, err
	}

	bi := &BoostIndexer{
		Network:        network,
		rpc:            rpc,
		es:             es,
		messageQueue:   messageQueue,
		filesDirectory: cfg.Share.Path,
		stop:           make(chan struct{}),
		interfaces:     interfaces,
		tokenViews:     tokenViews,
	}

	for _, opt := range opts {
		opt(bi)
	}

	if err := bi.init(); err != nil {
		return nil, err
	}

	return bi, nil
}

func (bi *BoostIndexer) init() error {
	if err := bi.es.CreateIndexes(); err != nil {
		return err
	}

	if bi.boost {
		if err := bi.fetchExternalProtocols(); err != nil {
			return err
		}
	}

	currentState, err := bi.es.GetLastBlock(bi.Network)
	if err != nil {
		return err
	}
	bi.state = currentState
	logger.Info("[%s] Current indexer state: %d", bi.Network, currentState.Level)

	currentProtocol, err := bi.es.GetProtocol(bi.Network, "", currentState.Level)
	if err != nil {
		header, err := bi.rpc.GetHeader(helpers.MaxInt64(1, currentState.Level))
		if err != nil {
			return err
		}
		currentProtocol, err = createProtocol(bi.es, bi.Network, header.Protocol, 0)
		if err != nil {
			return err
		}
	}
	bi.currentProtocol = currentProtocol
	logger.Info("[%s] Current network protocol: %s", bi.Network, currentProtocol.Hash)
	logger.Info("[%s] Data will be updated every %d seconds", bi.Network, currentProtocol.Constants.TimeBetweenBlocks)
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
		duration = 1 * time.Second
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-bi.stop:
			bi.stopped = true
			bi.messageQueue.Close()
			return
		case <-ticker.C:
			if err := bi.process(); err != nil {
				if err.Error() == "Same level" {
					if !everySecond {
						everySecond = true
						ticker.Stop()
						ticker = time.NewTicker(time.Duration(5) * time.Second)
					}
					continue
				}
				logger.Error(err)
				helpers.CatchErrorSentry(err)
			}

			if everySecond {
				everySecond = false
				ticker.Stop()
				ticker = time.NewTicker(duration)
			}
		}
	}
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
	for _, level := range levels {
		helpers.SetTagSentry("level", fmt.Sprintf("%d", level))

		select {
		case <-bi.stop:
			bi.stopped = true
			bi.messageQueue.Close()
			return errBcdQuit
		default:
		}

		currentHead, err := bi.rpc.GetHeader(level)
		if err != nil {
			return err
		}

		if bi.state.Level > 0 && currentHead.Predecessor != bi.state.Hash && !bi.boost {
			return errRollback
		}

		logger.Info("[%s] indexing %d block", bi.Network, level)

		if currentHead.Protocol != bi.currentProtocol.Hash {
			log.Printf("[%s] New protocol detected: %s -> %s", bi.Network, bi.currentProtocol.Hash, currentHead.Protocol)
			migrationModels, err := bi.migrate(currentHead)
			if err != nil {
				return err
			}
			if err := bi.saveModels(migrationModels); err != nil {
				return err
			}
		}

		parsedModels, err := bi.getDataFromBlock(bi.Network, currentHead)
		if err != nil {
			return err
		}
		parsedModels = append(parsedModels, bi.createBlock(currentHead))

		if err := bi.saveModels(parsedModels); err != nil {
			return err
		}
	}
	return nil
}

// Rollback -
func (bi *BoostIndexer) Rollback() error {
	logger.Warning("[%s] Rollback from %d", bi.Network, bi.state.Level)

	lastLevel, err := bi.getLastRollbackBlock()
	if err != nil {
		return err
	}

	if err := rollback.Rollback(bi.es, bi.messageQueue, bi.filesDirectory, bi.state, lastLevel); err != nil {
		return err
	}

	helpers.CatchErrorSentry(errors.Errorf("[%s] Rollback from %d to %d", bi.Network, bi.state.Level, lastLevel))

	newState, err := bi.es.GetLastBlock(bi.Network)
	if err != nil {
		return err
	}
	bi.state = newState
	logger.Info("[%s] New indexer state: %d", bi.Network, bi.state.Level)
	logger.Success("[%s] Rollback finished", bi.Network)
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

		block, err := bi.es.GetBlock(bi.Network, level)
		if err != nil {
			return 0, err
		}

		if block.Predecessor == headAtLevel.Predecessor {
			logger.Info("Found equal predecessors at level: %d", block.Level)
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

func (bi *BoostIndexer) validChainID(head noderpc.Header) bool {
	if bi.state.ChainID == "" {
		return bi.state.Level == 0
	}
	return bi.state.ChainID == head.ChainID
}

func (bi *BoostIndexer) process() error {
	head, err := bi.rpc.GetHead()
	if err != nil {
		return err
	}

	if !bi.validChainID(head) {
		return errors.Errorf("Invalid chain_id: %s (state) != %s (head)", bi.state.ChainID, head.ChainID)
	}

	logger.Info("[%s] Current node state: %d", bi.Network, head.Level)
	logger.Info("[%s] Current indexer state: %d", bi.Network, bi.state.Level)

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

		logger.Info("[%s] Found %d new levels", bi.Network, len(levels))

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
		logger.Success("[%s] Synced", bi.Network)
		return nil
	} else if head.Level < bi.state.Level {
		if err := bi.Rollback(); err != nil {
			return err
		}
	}

	return errors.Errorf("Same level")
}

func (bi *BoostIndexer) createBlock(head noderpc.Header) *models.Block {
	newBlock := models.Block{
		ID:          helpers.GenerateID(),
		Network:     bi.Network,
		Hash:        head.Hash,
		Predecessor: head.Predecessor,
		Protocol:    head.Protocol,
		ChainID:     head.ChainID,
		Level:       head.Level,
		Timestamp:   head.Timestamp,
	}

	bi.state = newBlock
	return &newBlock
}

func (bi *BoostIndexer) saveModels(items []elastic.Model) error {
	logger.Info("[%s] Found %d new models", bi.Network, len(items))
	if err := bi.es.BulkInsert(items); err != nil {
		return err
	}

	for i := range items {
		if err := bi.messageQueue.Send(items[i]); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) getDataFromBlock(network string, head noderpc.Header) ([]elastic.Model, error) {
	if head.Level <= 1 {
		return nil, nil
	}
	data, err := bi.rpc.GetOperations(head.Level)
	if err != nil {
		return nil, err
	}
	defaultParser := parsers.NewDefaultParser(bi.rpc, bi.es, bi.filesDirectory)
	defaultParser.SetConstants(bi.currentProtocol.Constants)
	defaultParser.SetInterface(bi.interfaces)
	defaultParser.SetTokenViews(bi.tokenViews)

	parsedModels := make([]elastic.Model, 0)
	for _, opg := range data.Array() {
		parsed, err := defaultParser.Parse(opg, network, head)
		if err != nil {
			return nil, err
		}
		parsedModels = append(parsedModels, parsed...)
	}

	return parsedModels, nil
}

func (bi *BoostIndexer) migrate(head noderpc.Header) ([]elastic.Model, error) {
	updates := make([]elastic.Model, 0)
	newModels := make([]elastic.Model, 0)

	if bi.currentProtocol.EndLevel == 0 && head.Level > 1 {
		logger.Info("[%s] Finalizing the previous protocol: %s", bi.Network, bi.currentProtocol.Alias)
		bi.currentProtocol.EndLevel = head.Level - 1
		updates = append(updates, &bi.currentProtocol)
	}

	newProtocol, err := bi.es.GetProtocol(bi.Network, head.Protocol, head.Level)
	if err != nil {
		logger.Warning("%s", err)
		newProtocol, err = createProtocol(bi.es, bi.Network, head.Protocol, head.Level)
		if err != nil {
			return nil, err
		}
	}

	if bi.Network == consts.Mainnet && head.Level == 1 {
		vestingMigrations, err := bi.vestingMigration(head)
		if err != nil {
			return nil, err
		}
		newModels = append(newModels, vestingMigrations...)
	} else {
		if bi.currentProtocol.SymLink == "" {
			return nil, errors.Errorf("[%s] Protocol should be initialized", bi.Network)
		}
		if newProtocol.SymLink != bi.currentProtocol.SymLink {
			migrations, migrationUpdates, err := bi.standartMigration(newProtocol)
			if err != nil {
				return nil, err
			}
			newModels = append(newModels, migrations...)
			if len(migrationUpdates) > 0 {
				updates = append(updates, migrationUpdates...)
			}
		} else {
			logger.Info("[%s] Same symlink %s for %s / %s",
				bi.Network, newProtocol.SymLink, bi.currentProtocol.Alias, newProtocol.Alias)
		}
	}

	bi.currentProtocol = newProtocol
	newModels = append(newModels, &newProtocol)

	if err := bi.es.BulkUpdate(updates); err != nil {
		return nil, err
	}

	logger.Info("[%s] Migration to %s is completed", bi.Network, bi.currentProtocol.Alias)
	return newModels, nil
}

func createProtocol(es elastic.IElastic, network, hash string, level int64) (protocol models.Protocol, err error) {
	logger.Info("[%s] Creating new protocol %s starting at %d", network, hash, level)
	protocol.SymLink, err = meta.GetProtoSymLink(hash)
	if err != nil {
		return
	}

	protocol.Alias = hash[:8]
	protocol.Network = network
	protocol.Hash = hash
	protocol.StartLevel = level
	protocol.ID = helpers.GenerateID()
	return
}

func (bi *BoostIndexer) standartMigration(newProtocol models.Protocol) ([]elastic.Model, []elastic.Model, error) {
	log.Printf("[%s] Try to find migrations...", bi.Network)
	contracts, err := bi.es.GetContracts(map[string]interface{}{
		"network": bi.Network,
	})
	if err != nil {
		return nil, nil, err
	}
	log.Printf("[%s] Now %d contracts are indexed", bi.Network, len(contracts))

	p := parsers.NewMigrationParser(bi.rpc, bi.es, bi.filesDirectory)
	newModels := make([]elastic.Model, 0)
	newUpdates := make([]elastic.Model, 0)
	for i := range contracts {
		logger.Info("Migrate %s...", contracts[i].Address)
		script, err := bi.rpc.GetScriptJSON(contracts[i].Address, newProtocol.StartLevel)
		if err != nil {
			return nil, nil, err
		}

		createdModels, updates, err := p.Parse(script, contracts[i], bi.currentProtocol, newProtocol)
		if err != nil {
			return nil, nil, err
		}

		if len(createdModels) > 0 {
			newModels = append(newModels, createdModels...)
		}
		if len(updates) > 0 {
			newUpdates = append(newUpdates, updates...)
		}
	}
	return newModels, newUpdates, nil
}

func (bi *BoostIndexer) vestingMigration(head noderpc.Header) ([]elastic.Model, error) {
	addresses, err := bi.rpc.GetContractsByBlock(head.Level)
	if err != nil {
		return nil, err
	}

	p := parsers.NewVestingParser(bi.rpc, bi.es, bi.filesDirectory, bi.interfaces)

	parsedModels := make([]elastic.Model, 0)
	for _, address := range addresses {
		if !strings.HasPrefix(address, "KT") {
			continue
		}

		data, err := bi.rpc.GetContractJSON(address, head.Level)
		if err != nil {
			return nil, err
		}

		parsed, err := p.Parse(data, head, bi.Network, address)
		if err != nil {
			return nil, err
		}
		parsedModels = append(parsedModels, parsed...)
	}

	return parsedModels, nil
}
