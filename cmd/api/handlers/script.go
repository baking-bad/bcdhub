package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

func (ctx *Context) getScript(network types.Network, address, symLink string) (*ast.Script, error) {
	data, err := ctx.getScriptBytes(network, address, symLink)
	if err != nil {
		return nil, err
	}
	return ast.NewScript(data)
}

func (ctx *Context) getScriptBytes(network types.Network, address, symLink string) ([]byte, error) {
	if symLink == "" {
		state, err := ctx.CachedCurrentBlock(network)
		if err != nil {
			return nil, err
		}
		symLink = state.Protocol.SymLink
	}
	return fetch.ContractBySymLink(network, address, symLink, ctx.SharePath)
}

func (ctx *Context) getParameterType(network types.Network, address, symLink string) (*ast.TypedAst, error) {
	script, err := ctx.getScript(network, address, symLink)
	if err != nil {
		return nil, err
	}
	return script.ParameterType()
}

func (ctx *Context) getStorageType(network types.Network, address, symLink string) (*ast.TypedAst, error) {
	script, err := ctx.getScript(network, address, symLink)
	if err != nil {
		return nil, err
	}
	return script.StorageType()
}
