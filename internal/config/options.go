package config

import (
	"errors"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/elastic"
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
	"github.com/go-pg/pg/v10"

	"github.com/baking-bad/bcdhub/internal/postgres/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/postgres/block"
	pgCore "github.com/baking-bad/bcdhub/internal/postgres/core"

	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// ContextOption -
type ContextOption func(ctx *Context)

// WithRPC -
func WithRPC(rpcConfig map[string]RPCConfig, cache bool) ContextOption {
	return func(ctx *Context) {
		if rpcProvider, ok := rpcConfig[ctx.Network.String()]; ok {
			if rpcProvider.URI == "" {
				return
			}
			opts := []noderpc.NodeOption{
				noderpc.WithTimeout(time.Second * time.Duration(rpcProvider.Timeout)),
			}
			if cache {
				opts = append(opts, noderpc.WithCache(ctx.Config.SharePath, ctx.Network.String()))
			}

			ctx.RPC = noderpc.NewPool([]string{rpcProvider.URI}, opts...)
		}
	}
}

// WithStorage -
func WithStorage(cfg StorageConfig, appName string, maxPageSize int64, maxConnCount, idleConnCount int) ContextOption {
	return func(ctx *Context) {
		if len(cfg.Elastic) == 0 {
			panic("Please set connection strings to storage in config")
		}
		defaultConn := pgCore.WaitNew(cfg.Postgres.ConnectionString(), appName, cfg.Timeout)
		defer defaultConn.Close()

		if result, err := defaultConn.DB.Exec(`SELECT datname FROM pg_catalog.pg_database WHERE datname = ?`, ctx.Network.String()); err != nil || result.RowsReturned() == 0 {
			if errors.Is(err, pg.ErrNoRows) || result.RowsReturned() == 0 {
				if _, err := defaultConn.DB.Exec("create database ?", pg.Ident(ctx.Network.String())); err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}

		networkConfig := PostgresConfig{
			Host:     cfg.Postgres.Host,
			Port:     cfg.Postgres.Port,
			User:     cfg.Postgres.User,
			Password: cfg.Postgres.Password,
			DBName:   ctx.Network.String(),
			SslMode:  cfg.Postgres.SslMode,
		}

		pg := pgCore.WaitNew(networkConfig.ConnectionString(), appName, cfg.Timeout,
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
		if svcCfg, ok := cfg[ctx.Network.String()]; ok {
			if svcCfg.MempoolURI == "" {
				return
			}
			ctx.Mempool = mempool.NewMempool(svcCfg.MempoolURI)
		}
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
