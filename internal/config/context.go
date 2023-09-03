package config

import (
	"github.com/baking-bad/bcdhub/internal/cache"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/services/mempool"
	"github.com/pkg/errors"
)

// Context -
type Context struct {
	Network types.Network
	RPC     noderpc.INode
	Mempool *mempool.Mempool

	StorageDB *core.Postgres

	Config     Config
	TzipSchema string

	Storage         models.GeneralRepository
	Accounts        account.Repository
	BigMapActions   bigmapaction.Repository
	BigMapDiffs     bigmapdiff.Repository
	Blocks          block.Repository
	Contracts       contract.Repository
	GlobalConstants contract.ConstantRepository
	Migrations      migration.Repository
	Operations      operation.Repository
	Protocols       protocol.Repository
	TicketUpdates   ticket.Repository
	Domains         domains.Repository
	Scripts         contract.ScriptRepository
	SmartRollups    smartrollup.Repository
	Partitions      *postgres.PartitionManager

	Cache *cache.Cache
}

// NewContext -
func NewContext(network types.Network, opts ...ContextOption) *Context {
	ctx := &Context{
		Network: network,
	}

	for _, opt := range opts {
		opt(ctx)
	}

	ctx.Cache = cache.NewCache(
		ctx.RPC, ctx.Accounts, ctx.Contracts, ctx.Protocols,
	)
	return ctx
}

// Close -
func (ctx *Context) Close() error {
	if ctx.StorageDB != nil {
		if err := ctx.StorageDB.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Contexts -
type Contexts map[types.Network]*Context

// NewContext -
func NewContexts(cfg Config, networks []string, opts ...ContextOption) Contexts {
	if len(networks) == 0 {
		panic("empty networks list in config file")
	}

	ctxs := make(Contexts)

	for i := range networks {
		networkType := types.NewNetwork(networks[i])
		if networkType == types.Empty {
			logger.Warning().Msgf("unknown network: %s", networks[i])
			continue
		}
		ctxs[networkType] = NewContext(networkType, opts...)
	}

	return ctxs
}

// Get -
func (ctxs Contexts) Get(network types.Network) (*Context, error) {
	if ctx, ok := ctxs[network]; ok {
		return ctx, nil
	}
	return nil, errors.Errorf("unknown network: %s", network.String())
}

// Any -
func (ctxs Contexts) Any() *Context {
	for _, ctx := range ctxs {
		return ctx
	}
	panic("empty contexts map")
}

// Close -
func (ctxs Contexts) Close() error {
	for _, c := range ctxs {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}
