package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/bigmap"
	"github.com/baking-bad/bcdhub/internal/postgres/contract"
	"github.com/baking-bad/bcdhub/internal/postgres/dapp"
	"github.com/baking-bad/bcdhub/internal/postgres/domains"
	"github.com/baking-bad/bcdhub/internal/postgres/migration"
	"github.com/baking-bad/bcdhub/internal/postgres/operation"
	"github.com/baking-bad/bcdhub/internal/postgres/protocol"
	"github.com/baking-bad/bcdhub/internal/postgres/service"
	"github.com/baking-bad/bcdhub/internal/postgres/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/postgres/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/postgres/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/postgres/transfer"
	"github.com/baking-bad/bcdhub/internal/postgres/tzip"

	"github.com/baking-bad/bcdhub/internal/postgres/block"
	pgCore "github.com/baking-bad/bcdhub/internal/postgres/core"

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
func WithStorage(cfg StorageConfig, appName string, maxPageSize int64) ContextOption {
	return func(ctx *Context) {
		if len(cfg.Elastic) == 0 {
			panic("Please set connection strings to storage in config")
		}

		pg := pgCore.WaitNew(cfg.Postgres, appName, cfg.Timeout, pgCore.WithPageSize(maxPageSize))
		ctx.StorageDB = pg
		ctx.Storage = pg
		ctx.BigMaps = bigmap.NewStorage(pg)
		ctx.BigMapActions = bigmap.NewActionStorage(pg)
		ctx.BigMapDiffs = bigmap.NewDiffStorage(pg)
		ctx.BigMapState = bigmap.NewStateStorage(pg)
		ctx.Blocks = block.NewStorage(pg)
		ctx.DApps = dapp.NewStorage(pg)
		ctx.Contracts = contract.NewStorage(pg)
		ctx.Migrations = migration.NewStorage(pg)
		ctx.Operations = operation.NewStorage(pg)
		ctx.Protocols = protocol.NewStorage(pg)
		ctx.TezosDomains = tezosdomain.NewStorage(pg)
		ctx.TokenBalances = tokenbalance.NewStorage(pg)
		ctx.TokenMetadata = tokenmetadata.NewStorage(pg)
		ctx.Transfers = transfer.NewStorage(pg)
		ctx.TZIP = tzip.NewStorage(pg)
		ctx.Domains = domains.NewStorage(pg)
		ctx.Services = service.NewStorage(pg)
	}
}

// WithSearch -
func WithSearch(cfg StorageConfig) ContextOption {
	return func(ctx *Context) {
		ctx.Searcher = elastic.WaitNew(cfg.Elastic, cfg.Timeout)
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
		svc := make(map[types.Network]tzkt.Service)
		for network, tzktProvider := range tzktConfig {
			typ := types.NewNetwork(network)
			svc[typ] = tzkt.NewServicesTzKT(network, tzktProvider.ServicesURI, time.Second*time.Duration(tzktProvider.Timeout))
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
		ctx.TezosDomainsContracts = make(map[types.Network]string)
		for network, address := range cfg {
			ctx.TezosDomainsContracts[types.NewNetwork(network)] = address
		}
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
