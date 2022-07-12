package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func getScript(contracts contract.Repository, address, symLink string) (*ast.Script, error) {
	data, err := getScriptBytes(contracts, address, symLink)
	if err != nil {
		return nil, err
	}
	return ast.NewScriptWithoutCode(data)
}

func getScriptBytes(contracts contract.Repository, address, symLink string) ([]byte, error) {
	script, err := contracts.Script(address, symLink)
	if err != nil {
		return nil, err
	}
	return script.Full()
}

func getParameterType(contracts contract.Repository, address, symLink string) (*ast.TypedAst, error) {
	return getScriptPart(contracts, address, symLink, consts.PARAMETER)
}

func getStorageType(contracts contract.Repository, address, symLink string) (*ast.TypedAst, error) {
	return getScriptPart(contracts, address, symLink, consts.STORAGE)
}

func getScriptPart(contracts contract.Repository, address, symLink, part string) (*ast.TypedAst, error) {
	data, err := contracts.ScriptPart(address, symLink, part)
	if err != nil {
		return nil, err
	}
	return ast.NewTypedAstFromBytes(data)
}

func getCurrentSymLink(blocks block.Repository) (string, error) {
	block, err := blocks.Last()
	if err != nil {
		return "", nil
	}
	return block.Protocol.SymLink, nil
}
