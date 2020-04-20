package indexer

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/cmd/indexer/parsers"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/index"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/rollback"
)

// BoostIndexer -
type BoostIndexer struct {
	Network         string
	UpdateTimer     int64
	rpc             noderpc.Pool
	es              *elastic.Elastic
	externalIndexer index.Indexer
	state           models.Block
	currentProtocol models.Protocol
	messageQueue    *mq.MQ
	filesDirectory  string
	boost           bool

	stop    chan struct{}
	stopped bool
}

// NewBoostIndexer -
func NewBoostIndexer(cfg Config, network string) (*BoostIndexer, error) {
	logger.Info("[%s] Creating indexer object...", network)
	config := cfg.Indexers[network]
	es := elastic.WaitNew([]string{cfg.Search.URI})
	rpc := noderpc.NewPool(config.RPC.URLs, time.Duration(config.RPC.Timeout)*time.Second)

	var externalIndexer index.Indexer
	var err error
	if config.Boost {
		externalIndexer, err = createExternalInexer(config.ExternalIndexer)
		if err != nil {
			return nil, err
		}
	}

	messageQueue, err := mq.New(cfg.Mq.URI, cfg.Mq.Queues)
	if err != nil {
		return nil, err
	}

	currentState, err := es.CurrentState(network)
	if err != nil {
		return nil, err
	}
	logger.Info("[%s] Current indexer state: %d", network, currentState.Level)

	currentProtocol, err := es.GetProtocol(network, currentState.Level)
	if err != nil {
		return nil, err
	}
	logger.Info("[%s] Current network protocol: %s", network, currentProtocol.Hash)

	logger.Info("[%s] Getting network constants...", network)
	constants, err := rpc.GetNetworkConstants()
	if err != nil {
		return nil, err
	}
	updateTimer := constants.Get("time_between_blocks.0").Int()
	logger.Info("[%s] Data will be updates every %d seconds", network, updateTimer)

	bi := &BoostIndexer{
		Network:         network,
		UpdateTimer:     updateTimer,
		rpc:             rpc,
		es:              es,
		externalIndexer: externalIndexer,
		messageQueue:    messageQueue,
		state:           currentState,
		currentProtocol: currentProtocol,
		filesDirectory:  cfg.FilesDirectory,
		boost:           config.Boost,
		stop:            make(chan struct{}),
	}
	err = bi.createIndexes()
	return bi, err
}

// Sync -
func (bi *BoostIndexer) Sync(wg *sync.WaitGroup) error {
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
		return nil
	}

	everySecond := false
	ticker := time.NewTicker(time.Duration(bi.UpdateTimer) * time.Second)
	for {
		select {
		case <-bi.stop:
			bi.stopped = true
			bi.messageQueue.Close()
			return nil
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
				ticker = time.NewTicker(time.Duration(bi.UpdateTimer) * time.Second)
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
		select {
		case <-bi.stop:
			bi.stopped = true
			bi.messageQueue.Close()
			return fmt.Errorf("bcd-quit")
		default:
		}

		currentHead, err := bi.rpc.GetHeader(level)
		if err != nil {
			return err
		}

		if currentHead.Predecessor != bi.state.Hash && !bi.boost {
			return fmt.Errorf("rollback")
		}

		logger.Info("[%s] indexing %d block", bi.Network, level)

		if currentHead.Protocol != bi.currentProtocol.Hash {
			log.Printf("[%s] New protocol detected: %s -> %s", bi.Network, bi.currentProtocol.Hash, currentHead.Protocol)
			if err := bi.migrate(currentHead); err != nil {
				return err
			}
		}

		operations, contracts, migrations, err := bi.getDataFromBlock(bi.Network, currentHead)
		if err != nil {
			return err
		}

		if len(contracts) > 0 {
			if err := bi.saveContracts(contracts); err != nil {
				return err
			}
		}
		if len(operations) > 0 {
			if err := bi.saveOperations(operations); err != nil {
				return err
			}
		}
		if len(migrations) > 0 {
			if err := bi.saveMigrations(migrations); err != nil {
				return err
			}
		}

		if err := bi.createAndSaveBlock(currentHead); err != nil {
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

	helpers.CatchErrorSentry(fmt.Errorf("[%s] Rollback from %d to %d", bi.Network, bi.state.Level, lastLevel))

	newState, err := bi.es.CurrentState(bi.Network)
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
	levels, err := bi.externalIndexer.GetContractOperationBlocks(bi.state.Level, head.Level)
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
		return fmt.Errorf("Invalid chain_id: %s (state) != %s (head)", bi.state.ChainID, head.ChainID)
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
			if strings.Contains(err.Error(), "bcd-quit") {
				return nil
			}
			if err.Error() == "rollback" {
				bi.Rollback()
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
		bi.Rollback()
	}

	return fmt.Errorf("Same level")
}

func (bi *BoostIndexer) getContracts() (map[string]struct{}, map[string]struct{}, error) {
	addresses, err := bi.es.GetContracts(map[string]interface{}{
		"network": bi.Network,
	})
	if err != nil {
		return nil, nil, err
	}
	res := make(map[string]struct{})
	spendable := make(map[string]struct{})
	for _, a := range addresses {
		res[a.Address] = struct{}{}
		if helpers.StringInArray(consts.SpendableTag, a.Tags) {
			spendable[a.Address] = struct{}{}
		}
	}

	return res, spendable, nil
}

func (bi *BoostIndexer) createAndSaveBlock(head noderpc.Header) error {
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

	if _, err := bi.es.AddDocumentWithID(newBlock, elastic.DocBlocks, newBlock.ID); err != nil {
		return err
	}

	bi.state = newBlock
	return nil
}

func (bi *BoostIndexer) saveContracts(contracts []models.Contract) error {
	logger.Info("[%s] Found %d new contracts", bi.Network, len(contracts))
	if err := bi.es.BulkInsertContracts(contracts); err != nil {
		return err
	}

	for j := range contracts {
		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueContracts, contracts[j].ID); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) saveOperations(operations []models.Operation) error {
	logger.Info("[%s] Found %d operations", bi.Network, len(operations))
	if err := bi.es.BulkInsertOperations(operations); err != nil {
		return err
	}

	for j := range operations {
		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueOperations, operations[j].ID); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) saveMigrations(migrations []models.Migration) error {
	logger.Info("[%s] Found %d migrations", bi.Network, len(migrations))
	if err := bi.es.BulkInsertMigrations(migrations); err != nil {
		return err
	}

	for j := range migrations {
		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueMigrations, migrations[j].ID); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) getDataFromBlock(network string, head noderpc.Header) ([]models.Operation, []models.Contract, []models.Migration, error) {
	data, err := bi.rpc.GetOperations(head.Level)
	if err != nil {
		return nil, nil, nil, err
	}
	defaultParser := parsers.NewDefaultParser(bi.rpc, bi.es, bi.filesDirectory)

	operations := make([]models.Operation, 0)
	contracts := make([]models.Contract, 0)
	migrations := make([]models.Migration, 0)
	for _, opg := range data.Array() {
		newOps, newContracts, newMigrations, err := defaultParser.Parse(opg, network, head)
		if err != nil {
			return nil, nil, nil, err
		}
		operations = append(operations, newOps...)
		contracts = append(contracts, newContracts...)
		migrations = append(migrations, newMigrations...)
	}

	return operations, contracts, migrations, nil
}

func (bi *BoostIndexer) migrate(head noderpc.Header) error {
	if bi.Network == consts.Mainnet && head.Level == 1 {
		if err := bi.vestingMigration(head); err != nil {
			return err
		}
	} else if bi.currentProtocol.Hash != "" {
		if err := bi.standartMigration(head); err != nil {
			return err
		}
	} else {
		return nil
	}

	if err := bi.updateProtocol(head); err != nil {
		return err
	}
	return nil
}

func (bi *BoostIndexer) updateProtocol(head noderpc.Header) error {
	protocol, err := bi.es.GetProtocol(bi.Network, head.Level)
	if err != nil {
		return err
	}
	bi.currentProtocol = protocol
	return nil
}

func (bi *BoostIndexer) standartMigration(head noderpc.Header) error {
	newSymLink, err := meta.GetProtoSymLink(head.Protocol)
	if err != nil {
		return err
	}
	if newSymLink == bi.currentProtocol.SymLink {
		return nil
	}

	log.Printf("[%s] Try to find migrations...", bi.Network)
	contracts, err := bi.es.GetContracts(map[string]interface{}{
		"network": bi.Network,
	})
	if err != nil {
		return err
	}
	log.Printf("[%s] Now %d contracts are indexed", bi.Network, len(contracts))

	p := parsers.NewMigrationParser(bi.rpc, bi.es, bi.filesDirectory)
	migrations := make([]models.Migration, 0)
	for i := range contracts {
		logger.Info("Migrate %s...", contracts[i].Address)
		script, err := bi.rpc.GetScriptJSON(contracts[i].Address, head.Level)
		if err != nil {
			return err
		}

		migration, err := p.Parse(script, head, contracts[i], bi.currentProtocol.Hash)
		if err != nil {
			return err
		}

		if migration != nil {
			migrations = append(migrations, *migration)
		}
	}
	if len(migrations) > 0 {
		if err := bi.saveMigrations(migrations); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) vestingMigration(head noderpc.Header) error {
	addresses, err := bi.rpc.GetContractsByBlock(head.Level)
	if err != nil {
		return err
	}

	p := parsers.NewVestingParser(bi.rpc, bi.es, bi.filesDirectory)

	migrations := make([]models.Migration, 0)
	contracts := make([]models.Contract, 0)
	for _, address := range addresses {
		if !strings.HasPrefix(address, "KT") {
			continue
		}

		data, err := bi.rpc.GetContractJSON(address, head.Level)
		if err != nil {
			return err
		}

		migration, contract, err := p.Parse(data, head, bi.Network, address)
		if err != nil {
			return err
		}
		migrations = append(migrations, migration)
		if contract != nil {
			contracts = append(contracts, *contract)
		}
	}

	if len(contracts) > 0 {
		if err := bi.saveContracts(contracts); err != nil {
			return err
		}
	}
	if len(migrations) > 0 {
		if err := bi.saveMigrations(migrations); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) createIndexes() error {
	for _, index := range []string{
		elastic.DocContracts,
		elastic.DocMetadata,
		elastic.DocBigMapDiff,
		elastic.DocOperations,
		elastic.DocBlocks,
		elastic.DocMigrations,
	} {
		if err := bi.es.CreateIndexIfNotExists(index); err != nil {
			return err
		}
	}
	return nil
}
