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
	s, err := es.CurrentState(network, models.StateOperation)
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
		index, err := createIndexer(es, cfg.Indexer, network, url.(string))
		if err != nil {
			return nil, err
		}
		idx[network] = index
	}
	return idx, nil
}

func syncIndexer(rpc *noderpc.NodeRPC, indexer index.Indexer, es *elastic.Elastic, network string) error {
	logger.Logf("-----------%s-----------", strings.ToUpper(network))
	cs, err := es.CurrentState(network, models.StateContract)
	if err != nil {
		return err
	}
	logger.Info("Current contract indexer state: %d", cs.Level)

	// Get current DB state
	s, ok := states[network]
	if !ok {
		return fmt.Errorf("Unknown network: %s", network)
	}
	logger.Info("Current state: %d", s.Level)

	if cs.Level > s.Level {
		addresses, err := es.GetContracts(map[string]interface{}{
			"network": network,
		})
		if err != nil {
			return err
		}

		levels, err := indexer.GetContractOperationBlocks(int(s.Level), addresses)
		if err != nil {
			return err
		}

		if len(levels) > 0 {
			logger.Info("Found %d contracts", len(addresses))
			logger.Info("Found %d new levels", len(levels))

			for _, l := range levels {
				ops, err := getOperations(rpc, es, l, network, addresses)
				if err != nil {
					return err
				}

				if s.Level < l {
					s.Level = l

					t, err := rpc.GetLevelTime(int(l))
					if err != nil {
						return err
					}
					s.Timestamp = t
				}
				if _, err = es.UpdateDoc(elastic.DocStates, s.ID, *s); err != nil {
					return err
				}

				logger.Info("[%d/%d] Found %d operations", l, cs.Level, len(ops))
				if len(ops) == 0 {
					continue
				}

				for j := range ops {
					if _, err := es.AddDocumentWithID(ops[j], elastic.DocOperations, ops[j].ID); err != nil {
						return err
					}
				}
			}
		}
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
