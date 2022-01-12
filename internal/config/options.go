package config

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/account"
	"github.com/baking-bad/bcdhub/internal/postgres/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/contract"
	cm "github.com/baking-bad/bcdhub/internal/postgres/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/postgres/dapp"
	"github.com/baking-bad/bcdhub/internal/postgres/domains"
	"github.com/baking-bad/bcdhub/internal/postgres/global_constant"
	"github.com/baking-bad/bcdhub/internal/postgres/migration"
	"github.com/baking-bad/bcdhub/internal/postgres/operation"
	"github.com/baking-bad/bcdhub/internal/postgres/protocol"
	"github.com/baking-bad/bcdhub/internal/postgres/service"
	"github.com/baking-bad/bcdhub/internal/postgres/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/postgres/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/postgres/transfer"
	"github.com/baking-bad/bcdhub/internal/services/mempool"

	"github.com/baking-bad/bcdhub/internal/postgres/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/postgres/block"
	pgCore "github.com/baking-bad/bcdhub/internal/postgres/core"

	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// ContextOption -
type ContextOption func(ctx *Context)

// WithRPC -
func WithRPC(rpcConfig map[string]RPCConfig) ContextOption {
	return func(ctx *Context) {
		if len(rpcConfig) == 0 {
			panic("RPC config is invalid")
		}
		rpc := make(map[types.Network]noderpc.INode)
		for name, rpcProvider := range rpcConfig {
			network := types.NewNetwork(name)
			rpc[network] = noderpc.NewPool(
				[]string{rpcProvider.URI},
				noderpc.WithTimeout(time.Second*time.Duration(rpcProvider.Timeout)),
			)
		}
		ctx.RPC = rpc
	}
}

// WithStorage -
func WithStorage(cfg StorageConfig, appName string, maxPageSize int64, maxConnCount, idleConnCount int) ContextOption {
	return func(ctx *Context) {
		if len(cfg.Elastic) == 0 {
			panic("Please set connection strings to storage in config")
		}

		pg := pgCore.WaitNew(cfg.Postgres, appName, cfg.Timeout,
			pgCore.WithPageSize(maxPageSize),
			pgCore.WithIdleConnections(idleConnCount),
			pgCore.WithMaxConnections(maxConnCount),
			// pgCore.WithQueryLogging(),
		)

		contractStorage := contract.NewStorage(pg)
		ctx.StorageDB = pg
		ctx.Storage = pg
		ctx.Accounts = account.NewStorage(pg)
		ctx.BigMapActions = bigmapaction.NewStorage(pg)
		ctx.Blocks = block.NewStorage(pg)
		ctx.BigMapDiffs = bigmapdiff.NewStorage(pg)
		ctx.DApps = dapp.NewStorage(pg)
		ctx.Contracts = contractStorage
		ctx.ContractMetadata = cm.NewStorage(pg)
		ctx.Migrations = migration.NewStorage(pg)
		ctx.Operations = operation.NewStorage(pg)
		ctx.Protocols = protocol.NewStorage(pg)
		ctx.TokenBalances = tokenbalance.NewStorage(pg)
		ctx.TokenMetadata = tokenmetadata.NewStorage(pg)
		ctx.Transfers = transfer.NewStorage(pg)
		ctx.GlobalConstants = global_constant.NewStorage(pg)
		ctx.Domains = domains.NewStorage(pg)
		ctx.Services = service.NewStorage(pg)
		ctx.Scripts = contractStorage
	}
}

// WithSearch -
func WithSearch(cfg StorageConfig) ContextOption {
	return func(ctx *Context) {
		searcher := elastic.WaitNew(cfg.Elastic, cfg.Timeout)
		ctx.Searcher = searcher
		ctx.Statistics = searcher
	}

}

// WithConfigCopy -
func WithConfigCopy(cfg Config) ContextOption {
	return func(ctx *Context) {
		ctx.Config = cfg
	}
}

// WithMempool -
func WithMempool(cfg map[string]ServiceConfig) ContextOption {
	return func(ctx *Context) {
		if len(cfg) == 0 {
			return
		}
		svc := make(map[types.Network]*mempool.Mempool)
		for network, svcCfg := range cfg {
			if svcCfg.MempoolURI == "" {
				continue
			}
			typ := types.NewNetwork(network)
			svc[typ] = mempool.NewMempool(svcCfg.MempoolURI)
		}
		ctx.MempoolServices = svc
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
		ctx.TezosDomainsContracts = make(map[types.Network]string)
		for network, address := range cfg {
			ctx.TezosDomainsContracts[types.NewNetwork(network)] = address
		}
	}
}
