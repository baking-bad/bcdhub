package config

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic/balanceupdate"
	"github.com/baking-bad/bcdhub/internal/elastic/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/elastic/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/elastic/block"
	"github.com/baking-bad/bcdhub/internal/elastic/bulk"
	"github.com/baking-bad/bcdhub/internal/elastic/contract"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/elastic/migration"
	"github.com/baking-bad/bcdhub/internal/elastic/operation"
	"github.com/baking-bad/bcdhub/internal/elastic/protocol"
	"github.com/baking-bad/bcdhub/internal/elastic/schema"
	"github.com/baking-bad/bcdhub/internal/elastic/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/elastic/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/elastic/transfer"
	"github.com/baking-bad/bcdhub/internal/elastic/tzip"
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

// WithElasticSearch -
func WithElasticSearch(esConfig ElasticSearchConfig) ContextOption {
	return func(ctx *Context) {
		es := core.WaitNew(esConfig.URI, esConfig.Timeout)

		ctx.Storage = es
		ctx.Bulk = bulk.NewStorage(es)
		ctx.BalanceUpdates = balanceupdate.NewStorage(es)
		ctx.BigMapActions = bigmapaction.NewStorage(es)
		ctx.BigMapDiffs = bigmapdiff.NewStorage(es)
		ctx.Blocks = block.NewStorage(es)
		ctx.Contracts = contract.NewStorage(es)
		ctx.Migrations = migration.NewStorage(es)
		ctx.Operations = operation.NewStorage(es)
		ctx.Protocols = protocol.NewStorage(es)
		ctx.Schema = schema.NewStorage(es)
		ctx.TezosDomains = tezosdomain.NewStorage(es)
		ctx.TokenBalances = tokenbalance.NewStorage(es)
		ctx.Transfers = transfer.NewStorage(es)
		ctx.TZIP = tzip.NewStorage(es)
	}
}

// WithDatabase -
func WithDatabase(dbConfig DatabaseConfig) ContextOption {
	return func(ctx *Context) {
		ctx.DB = database.WaitNew(dbConfig.ConnString, dbConfig.Timeout)
	}
}

// WithRabbit -
func WithRabbit(rabbitConfig RabbitConfig, service string, mqConfig MQConfig) ContextOption {
	return func(ctx *Context) {
		mqueues := make([]mq.Queue, 0)
		for name, params := range mqConfig.Queues {
			mqueues = append(mqueues, mq.Queue{
				Name:       name,
				AutoDelete: params.AutoDeleted,
				Durable:    !params.NonDurable,
			})
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
func WithLoadErrorDescriptions(filePath string) ContextOption {
	return func(ctx *Context) {
		if err := cerrors.LoadErrorDescriptions(filePath); err != nil {
			panic(err)
		}
	}
}

// WithContractsInterfaces -
func WithContractsInterfaces() ContextOption {
	return func(ctx *Context) {
		result, err := kinds.Load()
		if err != nil {
			panic(err)
		}
		ctx.Interfaces = result
	}
}

// WithAWS -
func WithAWS(cfg AWSConfig) ContextOption {
	return func(ctx *Context) {
		client, err := aws.New(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Region, cfg.BucketName)
		if err != nil {
			log.Panicf("aws client init error: %s", err)
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
