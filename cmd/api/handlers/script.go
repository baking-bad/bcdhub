package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
		state, err := ctx.Blocks.Last()
		if err != nil {
			return nil, err
		}
		symLink = state.Protocol.SymLink
	}
	script, err := ctx.Contracts.Script(address, symLink)
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

func getSymLink(network types.Network) string {
	if network == types.Jakartanet || network == types.Ithacanet {
		return bcd.SymLinkJakarta
	}
	return bcd.SymLinkBabylon
}
