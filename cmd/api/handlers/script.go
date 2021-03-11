package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
)

func (ctx *Context) getScript(address, network, protocol string) (*ast.Script, error) {
	data, err := ctx.getScriptBytes(address, network, protocol)
	if err != nil {
		return nil, err
	}
	return ast.NewScript(data)
}

func (ctx *Context) getScriptBytes(address, network, protocol string) ([]byte, error) {
	if protocol == "" {
		state, err := ctx.Blocks.Last(network)
		if err != nil {
			return nil, err
		}
		protocol = state.Protocol
	}
	return fetch.Contract(address, network, protocol, ctx.SharePath)
}

func (ctx *Context) getParameterType(address, network, protocol string) (*ast.TypedAst, error) {
	script, err := ctx.getScript(address, network, protocol)
	if err != nil {
		return nil, err
	}
	return script.ParameterType()
}

func (ctx *Context) getStorageType(address, network, protocol string) (*ast.TypedAst, error) {
	script, err := ctx.getScript(address, network, protocol)
	if err != nil {
		return nil, err
	}
	return script.StorageType()
}
