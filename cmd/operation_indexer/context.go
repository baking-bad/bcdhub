package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/index"
	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/mq"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

// Context -
type Context struct {
	MQ       *mq.MQ
	ES       *elastic.Elastic
	RPCs     map[string]noderpc.Pool
	Indexers map[string]index.Indexer
	States   map[string]*models.State
}

func newContext(cfg config) (*Context, error) {
	es := elastic.WaitNew([]string{cfg.Search.URI})

	states := make(map[string]*models.State)
	RPCs := createRPCs(cfg)
	indexers, err := createIndexers(es, cfg, states)
	if err != nil {
		return nil, err
	}

	messageQueue, err := mq.New(cfg.Mq.URI, cfg.Mq.Queues)
	if err != nil {
		return nil, err
	}

	return &Context{
		ES:       es,
		RPCs:     RPCs,
		Indexers: indexers,
		States:   states,
		MQ:       messageQueue,
	}, nil
}

// Close -
func (ctx *Context) Close() {
	ctx.MQ.Close()
}

func createRPCs(cfg config) map[string]noderpc.Pool {
	rpc := make(map[string]noderpc.Pool)
	for network, hosts := range cfg.NodeRPC {
		rpc[network] = noderpc.NewPool(hosts, time.Second*30)
	}
	return rpc
}

func createIndexer(es *elastic.Elastic, indexerType, network, url string, states map[string]*models.State) (index.Indexer, error) {
	if url == "" {
		return nil, nil
	}
	s, err := es.CurrentState(network, models.StateOperation)
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

func createIndexers(es *elastic.Elastic, cfg config, states map[string]*models.State) (map[string]index.Indexer, error) {
	idx := make(map[string]index.Indexer)
	indexerCfg := cfg.TzKT
	if cfg.Indexer == "tzstats" {
		indexerCfg = cfg.TzStats
	}
	for network, url := range indexerCfg {
		index, err := createIndexer(es, cfg.Indexer, network, url, states)
		if err != nil {
			return nil, err
		}
		idx[network] = index
	}
	return idx, nil
}

func (ctx *Context) getRPC(network string) (noderpc.Pool, error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return nil, fmt.Errorf("Unknown RPC network: %s", network)
	}
	return rpc, nil
}

func (ctx *Context) getIndexer(network string) (index.Indexer, error) {
	idx, ok := ctx.Indexers[network]
	if !ok {
		return nil, fmt.Errorf("Unknown RPC network: %s", network)
	}
	return idx, nil
}
