package main

import (
	"sync"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/index"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/mq"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/google/uuid"
)

func createContract(c index.Contract, rpc *noderpc.NodeRPC, es *elastic.Elastic, network, filesDirectory string) (n models.Contract, err error) {
	n.Level = c.Level
	n.Timestamp = c.Timestamp.UTC()
	n.Balance = c.Balance
	n.Address = c.Address
	n.Manager = c.Manager
	n.Delegate = c.Delegate
	n.Network = network

	n.ID = uuid.New().String()
	err = computeMetrics(rpc, es, &n, filesDirectory)
	return
}

func syncNetwork(ctx *Context, network string, wg *sync.WaitGroup) {
	defer wg.Done()

	rpc, err := ctx.GetRPC(network)
	if err != nil {
		logger.Errorf("[%s] %s", network, err.Error())
		return
	}

	indexer, err := ctx.GetIndexer(network)
	if err != nil {
		logger.Errorf("[%s] %s", network, err.Error())
		return
	}

	level, err := rpc.GetLevel()
	if err != nil {
		logger.Errorf("[%s] %s", network, err.Error())
		return
	}
	logger.Info("[%s] Current node state: %d", network, level)

	// Get current DB state
	s, ok := ctx.States[network]
	if !ok {
		logger.Errorf("Unknown network: %s", network)
		return
	}
	logger.Info("[%s] Current state: %d", network, s.Level)
	if level > s.Level {
		contracts, err := indexer.GetContracts(s.Level)
		if err != nil {
			logger.Errorf("[%s] %s", network, err.Error())
			return
		}
		logger.Info("[%s] New contracts: %d", network, len(contracts))

		if len(contracts) > 0 {
			for _, c := range contracts {
				n, err := createContract(c, rpc, ctx.ES, network, ctx.FilesDirectory)
				if err != nil {
					logger.Errorf("[%s %d] %s  [%s]", network, c.Level, err.Error(), c.Address)
					return
				}

				logger.Info("%s -> %s", network, n.Address)

				cID, err := ctx.ES.AddDocument(n, elastic.DocContracts)
				if err != nil {
					logger.Errorf("[%s] %s", network, err.Error())
					return
				}

				if err := ctx.MQ.Send(mq.ChannelNew, mq.QueueContracts, cID); err != nil {
					logger.Errorf("[%s] %s", network, err.Error())
					return
				}

				if s.Level < n.Level {
					s.Level = n.Level
					s.Timestamp = n.Timestamp
					s.Network = network
					s.Type = models.StateContract
				}

				if _, err = ctx.ES.UpdateDoc(elastic.DocStates, s.ID, s); err != nil {
					logger.Errorf("[%s] %s", network, err.Error())
					return
				}
			}
		}
		s.Level = level
		s.Timestamp = time.Now().UTC()
		if _, err = ctx.ES.UpdateDoc(elastic.DocStates, s.ID, s); err != nil {
			logger.Errorf("[%s] %s", network, err.Error())
			return
		}
		logger.Success("[%s] Synced", network)
	}
}

func process(ctx *Context) error {
	var wg sync.WaitGroup
	for network := range ctx.Indexers {
		wg.Add(1)
		go syncNetwork(ctx, network, &wg)
	}
	wg.Wait()
	return nil
}
