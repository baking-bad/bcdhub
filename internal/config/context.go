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
	cm "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/service"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/services/mempool"
	"github.com/microcosm-cc/bluemonday"
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

	Storage          models.GeneralRepository
	Accounts         account.Repository
	Statistics       models.Statistics
	BigMapActions    bigmapaction.Repository
	BigMapDiffs      bigmapdiff.Repository
	Blocks           block.Repository
	Contracts        contract.Repository
	DApps            dapp.Repository
	GlobalConstants  contract.ConstantRepository
	Migrations       migration.Repository
	Operations       operation.Repository
	Protocols        protocol.Repository
	TokenBalances    tokenbalance.Repository
	TokenMetadata    tokenmetadata.Repository
	Transfers        transfer.Repository
	ContractMetadata cm.Repository
	Domains          domains.Repository
	Services         service.Repository
	Scripts          contract.ScriptRepository

	Cache     *cache.Cache
	Sanitizer *bluemonday.Policy
}

// NewContext -
func NewContext(network types.Network, opts ...ContextOption) *Context {
	ctx := &Context{
		Sanitizer: bluemonday.UGCPolicy(),
		Network:   network,
	}
	ctx.Sanitizer.AllowAttrs("em")

	for _, opt := range opts {
		opt(ctx)
	}

	ctx.Cache = cache.NewCache(
		ctx.RPC, ctx.Accounts, ctx.Contracts, ctx.Protocols, ctx.ContractMetadata, ctx.Sanitizer,
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
