package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/index"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/google/uuid"
)

func createRPCs(cfg config) map[string]*noderpc.NodeRPC {
	rpc := make(map[string]*noderpc.NodeRPC)
	for i := range cfg.NodeRPC {
		nodeCfg := cfg.NodeRPC[i]
		rpc[nodeCfg.Network] = noderpc.NewNodeRPC(nodeCfg.Host)
		rpc[nodeCfg.Network].SetTimeout(time.Second * 30)
	}
	return rpc
}

func createIndexer(es *elastic.Elastic, indexerType, network, url string) (index.Indexer, error) {
	if url == "" {
		return nil, nil
	}
	s, err := es.CurrentState(network, models.StateContract)
	if err != nil {
		return nil, err
	}
	states[network] = &s

	logger.Info("[%s] Create %s indexer", network, indexerType)
	logger.Info("[%s] Current state %d level", network, s.Level)

	switch indexerType {
	case "tzkt":
		idx := index.NewTzKT(url, 30*time.Second)
		return idx, nil

	case "tzstats":
		idx := index.NewTzStats(url)
		return idx, nil
	default:
		log.Panicf("Unknown indexer type: %s", indexerType)
	}
	return nil, nil
}

func createIndexers(es *elastic.Elastic, cfg config) (map[string]index.Indexer, error) {
	idx := make(map[string]index.Indexer)
	indexerCfg := cfg.TzKT
	if cfg.Indexer == "tzstats" {
		indexerCfg = cfg.TzStats
	}
	for network, url := range indexerCfg {
		index, err := createIndexer(es, cfg.Indexer, network, url)
		if err != nil {
			return nil, err
		}
		idx[network] = index
	}
	return idx, nil
}

func createContract(c index.Contract, rpc *noderpc.NodeRPC, es *elastic.Elastic, network string) (n models.Contract, err error) {
	n.Level = c.Level
	n.Timestamp = c.Timestamp.UTC()
	n.Balance = c.Balance
	n.Address = c.Address
	n.Manager = c.Manager
	n.Delegate = c.Delegate
	n.Network = network

	n.ID = uuid.New().String()
	err = computeMetrics(rpc, es, &n)
	return
}

func syncIndexer(rpc *noderpc.NodeRPC, indexer index.Indexer, es *elastic.Elastic, network string, errChan chan error, done chan struct{}) {
	level, err := rpc.GetLevel()
	if err != nil {
		errChan <- err
		return
	}
	logger.Info("[%s] Current node state: %d", network, level)

	// Get current DB state
	s, ok := states[network]
	if !ok {
		errChan <- fmt.Errorf("Unknown network: %s", network)
		return
	}
	logger.Info("[%s] Current state: %d", network, s.Level)
	if level > s.Level {
		contracts, err := indexer.GetContracts(s.Level)
		if err != nil {
			errChan <- err
			return
		}
		logger.Info("[%s] New contracts: %d", network, len(contracts))

		if len(contracts) > 0 {
			for _, c := range contracts {
				n, err := createContract(c, rpc, es, network)
				if err != nil {
					errChan <- fmt.Errorf("[%s %d] %s  [%s]", network, c.Level, err.Error(), c.Address)
					return
				}

				logger.Info("%s -> %s", network, n.Address)

				if _, err := es.AddDocument(n, elastic.DocContracts); err != nil {
					errChan <- err
					return
				}

				if s.Level < n.Level {
					s.Level = n.Level
					s.Timestamp = n.Timestamp
					s.Network = network
					s.Type = models.StateContract
				}

				if _, err = es.UpdateDoc(elastic.DocStates, s.ID, s); err != nil {
					errChan <- err
					return
				}
			}
		}
		logger.Success("[%s] Synced", network)
	}
	done <- struct{}{}
}

func sync(rpcs map[string]*noderpc.NodeRPC, indexers map[string]index.Indexer, es *elastic.Elastic) error {
	errChan := make(chan error)
	done := make(chan struct{})
	for network, indexer := range indexers {
		rpc, ok := rpcs[network]
		if !ok {
			logger.Errorf("Unknown RPC network: %s", network)
			continue
		}

		go syncIndexer(rpc, indexer, es, network, errChan, done)
	}

	var count int
	for {
		select {
		case err := <-errChan:
			logger.Error(err)
			count++
		case <-done:
			count++
		}

		if count == len(rpcs) {
			close(errChan)
			close(done)
			return nil
		}
	}
}
