package config

import (
	"github.com/baking-bad/bcdhub/internal/aws"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/tzkt"
	"github.com/pkg/errors"
)

// Context -
type Context struct {
	DB           database.DB
	ES           elastic.IElastic
	MQ           *mq.QueueManager
	AWS          *aws.Client
	RPC          map[string]noderpc.INode
	TzKTServices map[string]*tzkt.ServicesTzKT

	Config    Config
	SharePath string

	Aliases    map[string]string
	Interfaces map[string]kinds.ContractKind
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
func (ctx *Context) GetTzKTService(network string) (*tzkt.ServicesTzKT, error) {
	if rpc, ok := ctx.TzKTServices[network]; ok {
		return rpc, nil
	}
	return nil, errors.Errorf("Unknown tzkt service network %s", network)
}

// LoadAliases -
func (ctx *Context) LoadAliases() error {
	if ctx.ES == nil {
		return errors.Errorf("Connection to database is not initialized")
	}
	aliases, err := ctx.ES.GetAliasesMap(consts.Mainnet)
	if err != nil {
		if !elastic.IsRecordNotFound(err) {
			return err
		}
		ctx.Aliases = make(map[string]string)
		return nil
	}
	ctx.Aliases = aliases
	return nil
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
