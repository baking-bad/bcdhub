package indexer

import (
	"fmt"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
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

	stop    chan struct{}
	stopped bool
}

// NewBoostIndexer -
func NewBoostIndexer(cfg Config, network string) (*BoostIndexer, error) {
	config := cfg.Indexers[network]
	es := elastic.WaitNew([]string{cfg.Search.URI})
	rpc := noderpc.NewPool(config.RPC.URLs, time.Duration(config.RPC.Timeout)*time.Second)
	if !config.Boost {
		return nil, fmt.Errorf("Invalid config: you have to set `boost` to true")
	}
	externalIndexer, err := createExternalInexer(config.ExternalIndexer)
	if err != nil {
		return nil, err
	}
	messageQueue, err := mq.New(cfg.Mq.URI, cfg.Mq.Queues)
	if err != nil {
		return nil, err
	}

	currentState, err := es.CurrentState(network, models.StateContract)
	if err != nil {
		return nil, err
	}

	return &BoostIndexer{
		Network:         network,
		UpdateTimer:     config.UpdateTimer,
		rpc:             rpc,
		es:              es,
		externalIndexer: externalIndexer,
		messageQueue:    messageQueue,
		state:           currentState,
		stop:            make(chan struct{}),
	}, nil
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

	ticker := time.NewTicker(time.Duration(bi.UpdateTimer) * time.Second)
	for {
		select {
		case <-bi.stop:
			bi.stopped = true
			bi.messageQueue.Close()
			return nil
		case <-ticker.C:
			if err := bi.process(); err != nil {
				logger.Error(err)
				helpers.CatchErrorSentry(err)
			}
		}
	}
}

// Stop -
func (bi *BoostIndexer) Stop() {
	bi.stop <- struct{}{}
}

func (bi *BoostIndexer) process() error {
	currentLevel, err := bi.rpc.GetLevel()
	if err != nil {
		return err
	}
	logger.Info("[%s] Current contract indexer state: %d", bi.Network, currentLevel)
	logger.Info("[%s] Current state: %d", bi.Network, bi.state.Level)

	if currentLevel > bi.state.Level {
		addresses, spendable, err := bi.getContracts()
		if err != nil {
			return err
		}

		levels, err := bi.externalIndexer.GetContractOperationBlocks(int(bi.state.Level), int(currentLevel), addresses, spendable)
		if err != nil {
			return err
		}

		if len(levels) == 0 {
			return nil
		}

		logger.Info("[%s] Found %d new levels", bi.Network, len(levels))

		for _, level := range levels {
			select {
			case <-bi.stop:
				bi.stopped = true
				bi.messageQueue.Close()
				return nil
			default:
			}

			ops, err := getOperations(bi.rpc, bi.es, level, bi.Network, addresses)
			if err != nil {
				return err
			}

			ts, err := bi.rpc.GetLevelTime(int(level))
			if err != nil {
				return err
			}

			logger.Info("[%s] %d/%d Found %d operations", bi.Network, level, currentLevel, len(ops))
			if err := bi.saveOperations(ops, ts); err != nil {
				return err
			}

			if err := bi.updateState(level, ts, &bi.state); err != nil {
				return err
			}
		}
	}

	logger.Success("[%s] Synced", bi.Network)
	return nil
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

func (bi *BoostIndexer) updateState(currentLevel int64, ts time.Time, s *models.State) error {
	if s.Level >= currentLevel {
		return nil
	}
	s.Level = currentLevel
	s.Timestamp = ts

	if _, err := bi.es.UpdateDoc(elastic.DocStates, s.ID, *s); err != nil {
		return err
	}
	return nil
}

func (bi *BoostIndexer) saveOperations(ops []models.Operation, ts time.Time) error {
	if len(ops) == 0 {
		return nil
	}

	for j := range ops {
		ops[j].Timestamp = ts
		if _, err := bi.es.AddDocumentWithID(ops[j], elastic.DocOperations, ops[j].ID); err != nil {
			return err
		}

		if err := bi.messageQueue.Send(mq.ChannelNew, mq.QueueOperations, ops[j].ID); err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}
