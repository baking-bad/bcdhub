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
)

// BoostIndexer -
type BoostIndexer struct {
	Network         string
	UpdateTimer     int64
	rpc             noderpc.Pool
	es              *elastic.Elastic
	externalIndexer index.Indexer
	state           models.State
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

	logger.Info("[%s] Getting current indexer state...", network)
	currentState, err := es.CurrentState(network)
	if err != nil {
		return nil, err
	}

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
			return nil
		default:
		}

		currentHead, err := bi.rpc.GetHeader(level)
		if err != nil {
			return err
		}

		logger.Info("[%s] indexing %d block", bi.Network, level)

		if currentHead.Protocol != bi.state.Protocol {
			log.Printf("[%s] New protocol detected: %s -> %s", bi.Network, bi.state.Protocol, currentHead.Protocol)
			if err := bi.migrate(currentHead); err != nil {
				return err
			}
		}

		operations, contracts, err := bi.getDataFromBlock(bi.Network, currentHead)
		if err != nil {
			return err
		}

		if len(contracts) > 0 {
			logger.Info("[%s] Found %d new contracts", bi.Network, len(contracts))
			if err := bi.saveContracts(contracts); err != nil {
				return err
			}
		}
		if len(operations) > 0 {
			logger.Info("[%s] Found %d operations", bi.Network, len(operations))
			if err := bi.saveOperations(operations); err != nil {
				return err
			}
		}

		if err := bi.updateState(currentHead); err != nil {
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) process() error {
	head, err := bi.rpc.GetHead()
	if err != nil {
		return err
	}
	logger.Info("[%s] Current node state: %d", bi.Network, head.Level)
	logger.Info("[%s] Current indexer state: %d", bi.Network, bi.state.Level)

	if head.Level > bi.state.Level {
		levels := make([]int64, 0)
		if bi.boost {
			levels, err = bi.externalIndexer.GetContractOperationBlocks(bi.state.Level, head.Level)
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
			return err
		}

		if err := bi.updateState(head); err != nil {
			return err
		}
		if bi.boost {
			bi.boost = false
		}
		logger.Success("[%s] Synced", bi.Network)
		return nil
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

func (bi *BoostIndexer) updateState(head noderpc.Header) error {
	if bi.state.Level >= head.Level {
		return nil
	}
	bi.state.Level = head.Level
	bi.state.Timestamp = head.Timestamp
	bi.state.Protocol = head.Protocol

	if _, err := bi.es.UpdateDoc(elastic.DocStates, bi.state.ID, bi.state); err != nil {
		return err
	}
	return nil
}

func (bi *BoostIndexer) saveContracts(contracts []models.Contract) error {
	if err := bi.es.BulkInsertContracts(contracts); err != nil {
		return err
	}

	for j := range contracts {
		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueContracts, contracts[j].ID); err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) saveOperations(ops []models.Operation) error {
	if err := bi.es.BulkInsertOperations(ops); err != nil {
		return err
	}

	for j := range ops {
		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueOperations, ops[j].ID); err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

func (bi *BoostIndexer) saveMigrations(migrations []models.Migration) error {
	return bi.es.BulkInsertMigrations(migrations)
}

func (bi *BoostIndexer) getDataFromBlock(network string, head noderpc.Header) ([]models.Operation, []models.Contract, error) {
	data, err := bi.rpc.GetOperations(head.Level)
	if err != nil {
		return nil, nil, err
	}
	defaultParser := parsers.NewDefaultParser(bi.rpc, bi.es, bi.filesDirectory)

	operations := make([]models.Operation, 0)
	contracts := make([]models.Contract, 0)
	for _, opg := range data.Array() {
		newOps, newContracts, err := defaultParser.Parse(opg, network, head)
		if err != nil {
			return nil, nil, err
		}
		operations = append(operations, newOps...)
		contracts = append(contracts, newContracts...)
	}

	return operations, contracts, nil
}

func (bi *BoostIndexer) migrate(head noderpc.Header) error {
	if bi.state.Protocol == "" {
		return bi.vestingMigration(head)
	}
	return bi.standartMigration(head)
}

func (bi *BoostIndexer) standartMigration(head noderpc.Header) error {
	newSymLink, err := meta.GetProtoSymLink(head.Protocol)
	if err != nil {
		return err
	}
	currentSymLink, err := meta.GetProtoSymLink(bi.state.Protocol)
	if err != nil {
		return err
	}
	if newSymLink == currentSymLink {
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
	for _, contract := range contracts {
		logger.Info("Migrate %s...", contract.Address)
		script, err := bi.rpc.GetScriptJSON(contract.Address, head.Level)
		if err != nil {
			return err
		}

		migration, err := p.Parse(script, head, contract)
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
	addresses, err := bi.rpc.GetContractsByBlock(1)
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

		data, err := bi.rpc.GetContractJSON(address, 1)
		if err != nil {
			return err
		}

		migration, contract, err := p.Parse(data, bi.Network, address, head.Protocol)
		if err != nil {
			return err
		}
		migrations = append(migrations, migration)
		if contract != nil {
			contracts = append(contracts, *contract)
		}
	}

	logger.Info("[%s] Found %d migrations", bi.Network, len(migrations))
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
		elastic.DocStates,
		elastic.DocMigrations,
	} {
		if err := bi.es.CreateIndexIfNotExists(index); err != nil {
			return err
		}
	}
	return nil
}
