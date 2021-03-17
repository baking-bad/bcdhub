package config

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/postgres/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/contract"
	"github.com/baking-bad/bcdhub/internal/postgres/migration"
	"github.com/baking-bad/bcdhub/internal/postgres/operation"
	"github.com/baking-bad/bcdhub/internal/postgres/protocol"
	"github.com/baking-bad/bcdhub/internal/postgres/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/postgres/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/postgres/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/postgres/transfer"
	"github.com/baking-bad/bcdhub/internal/postgres/tzip"
	"github.com/baking-bad/bcdhub/internal/reindexer"

	"github.com/baking-bad/bcdhub/internal/postgres/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/postgres/block"
	pgCore "github.com/baking-bad/bcdhub/internal/postgres/core"

	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/pinata"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

// ContextOption -
type ContextOption func(ctx *Context)

// WithRPC -
func WithRPC(rpcConfig map[string]RPCConfig) ContextOption {
	return func(ctx *Context) {
		if len(rpcConfig) == 0 {
			panic("RPC config is invalid")
		}
		rpc := make(map[string]noderpc.INode)
		for network, rpcProvider := range rpcConfig {
			rpc[network] = noderpc.NewPool(
				[]string{rpcProvider.URI},
				noderpc.WithTimeout(time.Second*time.Duration(rpcProvider.Timeout)),
			)
		}
		ctx.RPC = rpc
	}
}

// WithStorage -
func WithStorage(cfg StorageConfig, maxPageSize int64) ContextOption {
	return func(ctx *Context) {
		if len(cfg.Elastic) == 0 {
			panic("Please set connection strings to storage in config")
		}

		pg, err := pgCore.NewPostgres(cfg.Postgres)
		if err != nil {
			panic(err)
		}
		ctx.Storage = pg
		ctx.BigMapActions = bigmapaction.NewStorage(pg)
		ctx.Blocks = block.NewStorage(pg)
		ctx.BigMapDiffs = bigmapdiff.NewStorage(pg)
		ctx.Contracts = contract.NewStorage(pg)
		ctx.Migrations = migration.NewStorage(pg)
		ctx.Operations = operation.NewStorage(pg)
		ctx.Protocols = protocol.NewStorage(pg)
		ctx.TezosDomains = tezosdomain.NewStorage(pg)
		ctx.TokenBalances = tokenbalance.NewStorage(pg)
		ctx.TokenMetadata = tokenmetadata.NewStorage(pg)
		ctx.Transfers = transfer.NewStorage(pg)
		ctx.TZIP = tzip.NewStorage(pg)
	}
}

// WithDatabase -
func WithDatabase(dbConfig DatabaseConfig) ContextOption {
	return func(ctx *Context) {
		ctx.DB = database.WaitNew(dbConfig.ConnString, dbConfig.Timeout)
	}
}

// WithSearch -
func WithSearch(cfg StorageConfig) ContextOption {
	return func(ctx *Context) {
		if strings.HasPrefix(cfg.Elastic[0], "builtin://") {
			storage, err := reindexer.New(cfg.Elastic[0])
			if err != nil {
				panic(err)
			}
			ctx.Searcher = storage

			if err := ctx.Storage.CreateIndexes(); err != nil {
				panic(err)
			}
		} else {
			ctx.Searcher = elastic.WaitNew(cfg.Elastic, cfg.Timeout, 0)

		}
	}

}

// WithRabbit -
func WithRabbit(rabbitConfig RabbitConfig, service string, mqConfig MQConfig) ContextOption {
	return func(ctx *Context) {
		mqueues := make([]mq.Queue, 0)
		for name, params := range mqConfig.Queues {
			q := mq.Queue{
				Name:       name,
				AutoDelete: params.AutoDeleted,
				Durable:    !params.NonDurable,
			}

			if params.TTLSeconds > 0 {
				q.TTLSeconds = params.TTLSeconds
			}

			mqueues = append(mqueues, q)
		}

		ctx.MQ = mq.New(rabbitConfig.URI, service, mqConfig.NeedPublisher, rabbitConfig.Timeout, mqueues...)
	}
}

// WithShare -
func WithShare(path string) ContextOption {
	return func(ctx *Context) {
		if path == "" {
			panic("Empty share path in config")
		}
		ctx.SharePath = path
	}
}

// WithConfigCopy -
func WithConfigCopy(cfg Config) ContextOption {
	return func(ctx *Context) {
		ctx.Config = cfg
	}
}

// WithTzKTServices -
func WithTzKTServices(tzktConfig map[string]TzKTConfig) ContextOption {
	return func(ctx *Context) {
		if len(tzktConfig) == 0 {
			return
		}
		svc := make(map[string]tzkt.Service)
		for network, tzktProvider := range tzktConfig {
			svc[network] = tzkt.NewServicesTzKT(network, tzktProvider.ServicesURI, time.Second*time.Duration(tzktProvider.Timeout))
		}
		ctx.TzKTServices = svc
	}
}

// WithLoadErrorDescriptions -
func WithLoadErrorDescriptions() ContextOption {
	return func(ctx *Context) {
		if err := tezerrors.LoadErrorDescriptions(); err != nil {
			panic(err)
		}
	}
}

// WithAWS -
func WithAWS(cfg AWSConfig) ContextOption {
	return func(ctx *Context) {
		client, err := aws.New(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Region, cfg.BucketName)
		if err != nil {
			panic(fmt.Errorf("aws client init error: %s", err))
		}
		ctx.AWS = client
	}
}

// WithDomains -
func WithDomains(cfg TezosDomainsConfig) ContextOption {
	return func(ctx *Context) {
		ctx.Domains = cfg
	}
}

// WithPinata -
func WithPinata(cfg PinataConfig) ContextOption {
	return func(ctx *Context) {
		ctx.Pinata = pinata.New(cfg.Key, cfg.SecretKey, time.Second*time.Duration(cfg.TimeoutSeconds))
	}
}

// WithTzipSchema -
func WithTzipSchema(filePath string) ContextOption {
	return func(ctx *Context) {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		ctx.TzipSchema = string(data)
	}
}
