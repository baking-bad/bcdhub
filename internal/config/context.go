package config

import (
	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/balanceupdate"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/pinata"
	"github.com/baking-bad/bcdhub/internal/tzkt"
	"github.com/pkg/errors"
)

// Context -
type Context struct {
	DB           database.DB
	MQ           mq.Mediator
	AWS          *aws.Client
	RPC          map[string]noderpc.INode
	TzKTServices map[string]tzkt.Service
	Pinata       pinata.Service

	Config     Config
	SharePath  string
	TzipSchema string

	Domains map[string]string

	Storage        models.GeneralRepository
	BalanceUpdates balanceupdate.Repository
	BigMapActions  bigmapaction.Repository
	BigMapDiffs    bigmapdiff.Repository
	Blocks         block.Repository
	Contracts      contract.Repository
	Migrations     migration.Repository
	Operations     operation.Repository
	Protocols      protocol.Repository
	Schema         schema.Repository
	TezosDomains   tezosdomain.Repository
	TokenBalances  tokenbalance.Repository
	TokenMetadata  tokenmetadata.Repository
	Transfers      transfer.Repository
	TZIP           tzip.Repository
}

// NewContext -
func NewContext(opts ...ContextOption) *Context {
	ctx := &Context{}

	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

// GetRPC -
func (ctx *Context) GetRPC(network string) (noderpc.INode, error) {
	if rpc, ok := ctx.RPC[network]; ok {
		return rpc, nil
	}
	return nil, errors.Errorf("Unknown rpc network %s", network)
}

// GetTzKTService -
func (ctx *Context) GetTzKTService(network string) (tzkt.Service, error) {
	if rpc, ok := ctx.TzKTServices[network]; ok {
		return rpc, nil
	}
	return nil, errors.Errorf("Unknown tzkt service network %s", network)
}

// Close -
func (ctx *Context) Close() {
	if ctx.MQ != nil {
		ctx.MQ.Close()
	}
	if ctx.DB != nil {
		ctx.DB.Close()
	}
}
