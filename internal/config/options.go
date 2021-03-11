package config

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/elastic/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/elastic/block"
	"github.com/baking-bad/bcdhub/internal/elastic/contract"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/elastic/migration"
	"github.com/baking-bad/bcdhub/internal/elastic/operation"
	"github.com/baking-bad/bcdhub/internal/elastic/protocol"
	"github.com/baking-bad/bcdhub/internal/elastic/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/elastic/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/elastic/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/elastic/transfer"
	"github.com/baking-bad/bcdhub/internal/elastic/tzip"

	reindexerBMA "github.com/baking-bad/bcdhub/internal/reindexer/bigmapaction"
	reindexerBMD "github.com/baking-bad/bcdhub/internal/reindexer/bigmapdiff"
	reindexerBlock "github.com/baking-bad/bcdhub/internal/reindexer/block"
	reindexerContract "github.com/baking-bad/bcdhub/internal/reindexer/contract"
	reindexerCore "github.com/baking-bad/bcdhub/internal/reindexer/core"
	reindexerMigration "github.com/baking-bad/bcdhub/internal/reindexer/migration"
	reindexerOperation "github.com/baking-bad/bcdhub/internal/reindexer/operation"
	reindexerProtocol "github.com/baking-bad/bcdhub/internal/reindexer/protocol"
	reindexerTD "github.com/baking-bad/bcdhub/internal/reindexer/tezosdomain"
	reindexerTB "github.com/baking-bad/bcdhub/internal/reindexer/tokenbalance"
	reindexerTM "github.com/baking-bad/bcdhub/internal/reindexer/tokenmetadata"
	reindexerTransfer "github.com/baking-bad/bcdhub/internal/reindexer/transfer"
	reindexertzip "github.com/baking-bad/bcdhub/internal/reindexer/tzip"

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
func WithStorage(cfg StorageConfig) ContextOption {
	return func(ctx *Context) {
		if len(cfg.URI) == 0 {
			panic("Please set connection strings to storage in config")
		}
		if strings.HasPrefix(cfg.URI[0], "builtin://") {
			storage, err := reindexerCore.New(cfg.URI[0])
			if err != nil {
				panic(err)
			}

			ctx.Storage = storage
			ctx.BigMapActions = reindexerBMA.NewStorage(storage)
			ctx.BigMapDiffs = reindexerBMD.NewStorage(storage)
			ctx.Blocks = reindexerBlock.NewStorage(storage)
			ctx.Contracts = reindexerContract.NewStorage(storage)
			ctx.Migrations = reindexerMigration.NewStorage(storage)
			ctx.Operations = reindexerOperation.NewStorage(storage)
			ctx.Protocols = reindexerProtocol.NewStorage(storage)
			ctx.TezosDomains = reindexerTD.NewStorage(storage)
			ctx.TokenBalances = reindexerTB.NewStorage(storage)
			ctx.TokenMetadata = reindexerTM.NewStorage(storage)
			ctx.Transfers = reindexerTransfer.NewStorage(storage)
			ctx.TZIP = reindexertzip.NewStorage(storage)

			if err := ctx.Storage.CreateIndexes(); err != nil {
				panic(err)
			}
		} else {
			es := core.WaitNew(cfg.URI, cfg.Timeout)

			ctx.Storage = es
			ctx.BigMapActions = bigmapaction.NewStorage(es)
			ctx.BigMapDiffs = bigmapdiff.NewStorage(es)
			ctx.Blocks = block.NewStorage(es)
			ctx.Contracts = contract.NewStorage(es)
			ctx.Migrations = migration.NewStorage(es)
			ctx.Operations = operation.NewStorage(es)
			ctx.Protocols = protocol.NewStorage(es)
			ctx.TezosDomains = tezosdomain.NewStorage(es)
			ctx.TokenBalances = tokenbalance.NewStorage(es)
			ctx.TokenMetadata = tokenmetadata.NewStorage(es)
			ctx.Transfers = transfer.NewStorage(es)
			ctx.TZIP = tzip.NewStorage(es)
		}
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
