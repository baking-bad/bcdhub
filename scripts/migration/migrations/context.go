package migrations

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/index"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Context -
type Context struct {
	MQ       *mq.MQ
	ES       *elastic.Elastic
	RPCs     map[string]noderpc.Pool
	Indexers map[string]index.Indexer
	DB       database.DB

	Config Config
}

// NewContext - creates migration context from config
func NewContext(cfg Config) (*Context, error) {
	es, err := elastic.New([]string{cfg.Search.URI})
	if err != nil {
		return nil, err
	}
	if err := meta.LoadProtocols(es); err != nil {
		return nil, err
	}
	RPCs := createRPCs(cfg)
	indexers, err := createIndexers(es, cfg)
	if err != nil {
		return nil, err
	}

	messageQueue, err := mq.New(cfg.Mq.URI, cfg.Mq.Queues)
	if err != nil {
		return nil, err
	}

	dbUser := os.Getenv("POSTGRES_USER")
	if dbUser == "" {
		return nil, fmt.Errorf("POSTGRES_USER env is not set")
	}
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD env is not set")
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		return nil, fmt.Errorf("POSTGRES_DB env is not set")
	}

	dbURI := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s", cfg.DB.Host, cfg.DB.Port, dbUser, dbName, dbPassword, cfg.DB.SSLMode)
	db, err := database.New(dbURI)
	if err != nil {
		return nil, err
	}

	return &Context{
		ES:       es,
		RPCs:     RPCs,
		Indexers: indexers,
		MQ:       messageQueue,
		DB:       db,
		Config:   cfg,
	}, nil
}

// Close -
func (ctx *Context) Close() {
	ctx.MQ.Close()
	ctx.DB.Close()
}

func createRPCs(cfg Config) map[string]noderpc.Pool {
	rpc := make(map[string]noderpc.Pool)
	for network, hosts := range cfg.NodeRPC {
		rpc[network] = noderpc.NewPool(hosts, time.Second*30)
	}
	return rpc
}

func createIndexer(es *elastic.Elastic, indexerType, network, url string) (index.Indexer, error) {
	if url == "" {
		return nil, nil
	}

	logger.Info("[%s] Create %s indexer", network, indexerType)

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

func createIndexers(es *elastic.Elastic, cfg Config) (map[string]index.Indexer, error) {
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

// GetRPC -
func (ctx *Context) GetRPC(network string) (noderpc.Pool, error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return nil, fmt.Errorf("Unknown RPC network: %s", network)
	}
	return rpc, nil
}

// GetIndexer -
func (ctx *Context) GetIndexer(network string) (index.Indexer, error) {
	idx, ok := ctx.Indexers[network]
	if !ok {
		return nil, fmt.Errorf("Unknown RPC network: %s", network)
	}
	return idx, nil
}
