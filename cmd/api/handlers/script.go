package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

func (ctx *Context) getScript(network types.Network, address, protocol string) (*ast.Script, error) {
	data, err := ctx.getScriptBytes(network, address, protocol)
	if err != nil {
		return nil, err
	}
	return ast.NewScript(data)
}

func (ctx *Context) getScriptBytes(network types.Network, address, protocol string) ([]byte, error) {
	if protocol == "" {
		state, err := ctx.CachedCurrentBlock(network)
		if err != nil {
			return nil, err
		}
		protocol = state.Protocol
	}
	return fetch.Contract(network, address, protocol, ctx.SharePath)
}

func (ctx *Context) getParameterType(network types.Network, address, protocol string) (*ast.TypedAst, error) {
	script, err := ctx.getScript(network, address, protocol)
	if err != nil {
		return nil, err
	}
	return script.ParameterType()
}

func (ctx *Context) getStorageType(network types.Network, address, protocol string) (*ast.TypedAst, error) {
	script, err := ctx.getScript(network, address, protocol)
	if err != nil {
		return nil, err
	}
	return script.StorageType()
}
