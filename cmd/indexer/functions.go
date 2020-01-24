package main

import (
	"fmt"
	"log"
	"strings"
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

	logger.Info("Create %s %s indexer", indexerType, network)
	logger.Info("Current state %d level", s.Level)

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

func syncIndexer(rpc *noderpc.NodeRPC, indexer index.Indexer, es *elastic.Elastic, network string) error {
	logger.Logf("-----------%s-----------", strings.ToUpper(network))
	level, err := rpc.GetLevel()
	if err != nil {
		return err
	}
	logger.Info("Current node state: %d", level)

	// Get current DB state
	s, ok := states[network]
	if !ok {
		return fmt.Errorf("Unknown network: %s", network)
	}
	logger.Info("Current state: %d", s.Level)
	if level > s.Level {
		contracts, err := indexer.GetContracts(s.Level)
		if err != nil {
			return err
		}
		logger.Info("New contracts: %d", len(contracts))

		if len(contracts) > 0 {
			for _, c := range contracts {
				n, err := createContract(c, rpc, es, network)
				if err != nil {
					logger.Error(err)
					continue
				}

				logger.Info("[%s] Contract found", n.Address)

				if _, err := es.AddDocument(n, elastic.DocContracts); err != nil {
					return err
				}

				if s.Level < n.Level {
					s.Level = n.Level
					s.Timestamp = n.Timestamp
					s.Network = network
					s.Type = models.StateContract
				}

				if _, err = es.UpdateDoc(elastic.DocStates, s.ID, s); err != nil {
					return err
				}
			}
		}
		logger.Success("[%s] Synced", network)
	}
	return nil
}

func sync(rpcs map[string]*noderpc.NodeRPC, indexers map[string]index.Indexer, es *elastic.Elastic) error {
	for network, indexer := range indexers {
		rpc, ok := rpcs[network]
		if !ok {
			logger.Errorf("Unknown RPC network: %s", network)
			continue
		}

		if err := syncIndexer(rpc, indexer, es, network); err != nil {
			logger.Error(err)
			continue
		}
	}
	return nil
}
