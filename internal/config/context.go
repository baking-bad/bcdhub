package config

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

// Context -
type Context struct {
	DB           database.DB
	ES           *elastic.Elastic
	MQ           *mq.MQ
	RPC          map[string]noderpc.INode
	TzKTServices map[string]*tzkt.ServicesTzKT

	Config    Config
	SharePath string

	Aliases map[string]string
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
	return nil, fmt.Errorf("Unknown rpc network %s", network)
}

// GetTzKTService -
func (ctx *Context) GetTzKTService(network string) (*tzkt.ServicesTzKT, error) {
	if rpc, ok := ctx.TzKTServices[network]; ok {
		return rpc, nil
	}
	return nil, fmt.Errorf("Unknown tzkt service network %s", network)
}

// LoadAliases -
func (ctx *Context) LoadAliases() error {
	if ctx.DB == nil {
		return fmt.Errorf("Connection to database is not initialized")
	}
	aliases, err := ctx.DB.GetAliasesMap(consts.Mainnet)
	if err != nil {
		return err
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
