package handlers

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/cache"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func getScript(ctx context.Context, cache *cache.Cache, address, symLink string) (*ast.Script, error) {
	data, err := getScriptBytes(ctx, cache, address, symLink)
	if err != nil {
		return nil, err
	}
	return ast.NewScriptWithoutCode(data)
}

func getScriptBytes(ctx context.Context, cache *cache.Cache, address, symLink string) ([]byte, error) {
	script, err := cache.Script(ctx, address, symLink)
	if err != nil {
		return nil, err
	}
	return script.Full()
}

func getParameterType(ctx context.Context, contracts contract.Repository, address, symLink string) (*ast.TypedAst, error) {
	return getScriptPart(ctx, contracts, address, symLink, consts.PARAMETER)
}

func getStorageType(ctx context.Context, contracts contract.Repository, address, symLink string) (*ast.TypedAst, error) {
	return getScriptPart(ctx, contracts, address, symLink, consts.STORAGE)
}

func getScriptPart(ctx context.Context, contracts contract.Repository, address, symLink, part string) (*ast.TypedAst, error) {
	data, err := contracts.ScriptPart(ctx, address, symLink, part)
	if err != nil {
		return nil, err
	}
	return ast.NewTypedAstFromBytes(data)
}

func getCurrentSymLink(ctx context.Context, blocks block.Repository) (string, error) {
	block, err := blocks.Last(ctx)
	if err != nil {
		return "", nil
	}
	return block.Protocol.SymLink, nil
}
