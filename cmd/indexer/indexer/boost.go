package indexer

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/cmd/indexer/parsers"
	"github.com/baking-bad/bcdhub/internal/config"
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

<<<<<<< HEAD
func (bi *BoostIndexer) fetchExternalProtocols() error {
	logger.Info("[%s] Fetching external protocols", bi.Network)
	existingProtocols, err := bi.es.GetProtocolsByNetwork(bi.Network)
	if err != nil {
=======
func fetchExternalProtocols(es *elastic.Elastic, externalIndexer index.Indexer, network string) error {
	logger.Info("[%s] Fetching external protocols", network)
	var existingProtocols []models.Protocol
	if err := es.GetByNetwork(network, &existingProtocols); err != nil {
>>>>>>> Refactor getAll methods
		return err
	}

	exists := make(map[string]bool, 0)
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
		protocols = append(protocols, &models.Protocol{
			ID:         helpers.GenerateID(),
			Hash:       extProtocols[i].Hash,
			Alias:      alias,
			StartLevel: extProtocols[i].StartLevel,
			EndLevel:   extProtocols[i].LastLevel,
			SymLink:    symLink,
			Network:    bi.Network,
		})
		log.Printf("[%s] Fetched %s", bi.Network, alias)
	}

<<<<<<< HEAD
	if len(protocols) > 0 {
		return bi.es.BulkInsertProtocols(protocols)
	}
	return nil
=======
	return es.BulkInsert(protocols)
>>>>>>> Refactor getAll methods
}

// NewBoostIndexer -
func NewBoostIndexer(cfg config.Config, network string, opts ...BoostIndexerOption) (*BoostIndexer, error) {
	logger.Info("[%s] Creating indexer object...", network)
	es := elastic.WaitNew([]string{cfg.Elastic.URI})
	rpcProvider, ok := cfg.RPC[network]
	if !ok {
		return nil, fmt.Errorf("Unknown network %s", network)
	}
	rpc := noderpc.NewPool([]string{rpcProvider.URI}, time.Duration(rpcProvider.Timeout)*time.Second)

	messageQueue, err := mq.New(cfg.RabbitMQ.URI, cfg.RabbitMQ.Queues)
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

	currentState, err := bi.es.CurrentState(bi.Network)
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

	logger.Info("[%s] Getting network constants...", bi.Network)
	constants, err := bi.rpc.GetNetworkConstants()
	if err != nil {
		return err
	}
	bi.UpdateTimer = constants.Get("time_between_blocks.0").Int()
	logger.Info("[%s] Data will be updated every %d seconds", bi.Network, bi.UpdateTimer)
	return nil
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

		if bi.state.Level > 0 && currentHead.Predecessor != bi.state.Hash && !bi.boost {
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

		if err := bi.saveContracts(contracts); err != nil {
			return err
		}
		if err := bi.saveOperations(operations); err != nil {
			return err
		}
		if err := bi.saveMigrations(migrations); err != nil {
			return err
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

func (bi *BoostIndexer) saveContracts(contracts []elastic.Model) error {
	logger.Info("[%s] Found %d new contracts", bi.Network, len(contracts))
	if err := bi.es.BulkInsert(contracts); err != nil {
		return err
	}

	for j := range contracts {
		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueContracts, contracts[j].GetID()); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) saveOperations(operations []elastic.Model) error {
	logger.Info("[%s] Found %d operations", bi.Network, len(operations))
	if err := bi.es.BulkInsert(operations); err != nil {
		return err
	}

	for j := range operations {
		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueOperations, operations[j].GetID()); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) saveMigrations(migrations []elastic.Model) error {
	logger.Info("[%s] Found %d migrations", bi.Network, len(migrations))
	if err := bi.es.BulkInsert(migrations); err != nil {
		return err
	}

	for j := range migrations {
		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueMigrations, migrations[j].GetID()); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) getDataFromBlock(network string, head noderpc.Header) ([]elastic.Model, []elastic.Model, []elastic.Model, error) {
	data, err := bi.rpc.GetOperations(head.Level)
	if err != nil {
		return nil, nil, nil, err
	}
	defaultParser := parsers.NewDefaultParser(bi.rpc, bi.es, bi.filesDirectory)

	operations := make([]elastic.Model, 0)
	contracts := make([]elastic.Model, 0)
	migrations := make([]elastic.Model, 0)
	for _, opg := range data.Array() {
		newOps, newContracts, newMigrations, err := defaultParser.Parse(opg, network, head)
		if err != nil {
			return nil, nil, nil, err
		}
		for i := range newOps {
			operations = append(operations, newOps[i])
		}
		for i := range newContracts {
			contracts = append(contracts, newContracts[i])
		}
		for i := range newMigrations {
			migrations = append(migrations, newMigrations[i])
		}
	}

	return operations, contracts, migrations, nil
}

func (bi *BoostIndexer) migrate(head noderpc.Header) error {
	if bi.currentProtocol.EndLevel == 0 && head.Level > 1 {
		logger.Info("[%s] Finalizing the previous protocol: %s", bi.Network, bi.currentProtocol.Alias)
		bi.currentProtocol.EndLevel = head.Level - 1
		_, err := bi.es.UpdateDoc(elastic.DocProtocol, bi.currentProtocol.ID, bi.currentProtocol)
		if err != nil {
			return err
		}
	}

	newProtocol, err := bi.es.GetProtocol(bi.Network, head.Protocol, head.Level)
	if err != nil {
		logger.Warning("%s", err)
		newProtocol, err = createProtocol(bi.es, bi.Network, head.Protocol, head.Level)
		if err != nil {
			return err
		}
	}

	if bi.Network == consts.Mainnet && head.Level == 1 {
		if err := bi.vestingMigration(head); err != nil {
			return err
		}
	} else {
		if bi.currentProtocol.SymLink == "" {
			return fmt.Errorf("[%s] Protocol should be initialized", bi.Network)
		}
		if newProtocol.SymLink != bi.currentProtocol.SymLink {
			if err := bi.standartMigration(newProtocol); err != nil {
				return err
			}
		} else {
			logger.Info("[%s] Same symlink %s for %s / %s",
				bi.Network, newProtocol.SymLink, bi.currentProtocol.Alias, newProtocol.Alias)
		}
	}

	bi.currentProtocol = newProtocol
	logger.Info("[%s] Migration to %s is completed", bi.Network, bi.currentProtocol.Alias)
	return nil
}

func createProtocol(es *elastic.Elastic, network, hash string, level int64) (protocol models.Protocol, err error) {
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

	_, err = es.AddDocumentWithID(protocol, elastic.DocProtocol, protocol.ID)
	if err != nil {
		return
	}
	return
}

func (bi *BoostIndexer) standartMigration(newProtocol models.Protocol) error {
	log.Printf("[%s] Try to find migrations...", bi.Network)
	contracts, err := bi.es.GetContracts(map[string]interface{}{
		"network": bi.Network,
	})
	if err != nil {
		return err
	}
	log.Printf("[%s] Now %d contracts are indexed", bi.Network, len(contracts))

	p := parsers.NewMigrationParser(bi.rpc, bi.es, bi.filesDirectory)
	migrations := make([]elastic.Model, 0)
	for i := range contracts {
		logger.Info("Migrate %s...", contracts[i].Address)
		script, err := bi.rpc.GetScriptJSON(contracts[i].Address, newProtocol.StartLevel)
		if err != nil {
			return err
		}

		migration, err := p.Parse(script, contracts[i], bi.currentProtocol, newProtocol)
		if err != nil {
			return err
		}

		if migration != nil {
			migrations = append(migrations, migration)
		}
	}
	if err := bi.saveMigrations(migrations); err != nil {
		return err
	}
	return nil
}

func (bi *BoostIndexer) vestingMigration(head noderpc.Header) error {
	addresses, err := bi.rpc.GetContractsByBlock(head.Level)
	if err != nil {
		return err
	}

	p := parsers.NewVestingParser(bi.rpc, bi.es, bi.filesDirectory)

	migrations := make([]elastic.Model, 0)
	contracts := make([]elastic.Model, 0)
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
			contracts = append(contracts, contract)
		}
	}

	if err := bi.saveContracts(contracts); err != nil {
		return err
	}
	if err := bi.saveMigrations(migrations); err != nil {
		return err
	}
	return nil
}
