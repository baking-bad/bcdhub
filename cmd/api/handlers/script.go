package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
)

func getScript(ctx *config.Context, address, symLink string) (*ast.Script, error) {
	data, err := ctx.Cache.ScriptBytes(address, symLink)
	if err != nil {
		return nil, err
	}
	return ast.NewScriptWithoutCode(data)
}

func getScriptBytes(ctx *config.Context, address, symLink string) ([]byte, error) {
	if symLink == "" {
		state, err := ctx.Cache.CurrentBlock()
		if err != nil {
			return nil, err
		}
		symLink = state.Protocol.SymLink
	}
	script, err := ctx.Contracts.Script(ctx.Network, address, symLink)
	if err != nil {
		return nil, err
	}
	return script.Full()
}

func getParameterType(ctx *config.Context, address, symLink string) (*ast.TypedAst, error) {
	script, err := getScript(ctx, address, symLink)
	if err != nil {
		return nil, err
	}
	return script.ParameterType()
}

func getStorageType(ctx *config.Context, address, symLink string) (*ast.TypedAst, error) {
	script, err := getScript(ctx, address, symLink)
	if err != nil {
		return nil, err
	}
	return script.StorageType()
}
