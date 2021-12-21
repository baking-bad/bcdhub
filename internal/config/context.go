package config

import (
	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/cache"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/service"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/baking-bad/bcdhub/internal/services/mempool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
)

// Context -
type Context struct {
	AWS             *aws.Client
	RPC             map[types.Network]noderpc.INode
	MempoolServices map[types.Network]*mempool.Mempool

	StorageDB *core.Postgres

	Config     Config
	TzipSchema string

	TezosDomainsContracts map[types.Network]string

	Storage         models.GeneralRepository
	Statistics      models.Statistics
	BigMapActions   bigmapaction.Repository
	BigMapDiffs     bigmapdiff.Repository
	Blocks          block.Repository
	Contracts       contract.Repository
	DApps           dapp.Repository
	GlobalConstants global_constant.Repository
	Migrations      migration.Repository
	Operations      operation.Repository
	Protocols       protocol.Repository
	TokenBalances   tokenbalance.Repository
	TokenMetadata   tokenmetadata.Repository
	Transfers       transfer.Repository
	TZIP            tzip.Repository
	Domains         domains.Repository
	Services        service.Repository
	Scripts         contract.ScriptRepository

	Searcher search.Searcher

	Cache     *cache.Cache
	Sanitizer *bluemonday.Policy
}

// NewContext -
func NewContext(opts ...ContextOption) *Context {
	ctx := &Context{
		Cache:     cache.NewCache(),
		Sanitizer: bluemonday.UGCPolicy(),
	}
	ctx.Sanitizer.AllowAttrs("em")

	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

// GetRPC -
func (ctx *Context) GetRPC(network types.Network) (noderpc.INode, error) {
	if rpc, ok := ctx.RPC[network]; ok {
		return rpc, nil
	}
	return nil, errors.Errorf("unknown rpc: %s", network)
}

// GetMempoolService -
func (ctx *Context) GetMempoolService(network types.Network) (*mempool.Mempool, error) {
	if rpc, ok := ctx.MempoolServices[network]; ok {
		return rpc, nil
	}
	return nil, errors.Errorf("unknown mempool service: %s", network)
}

// Close -
func (ctx *Context) Close() {
	if ctx.StorageDB != nil {
		ctx.StorageDB.Close()
	}
}
