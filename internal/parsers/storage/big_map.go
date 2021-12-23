package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

// errors -
var (
	ErrBigMapNotFound = errors.New("Big map is not found")
)

// GetBigMapPtr -
func GetBigMapPtr(repo models.GeneralRepository, contracts contract.Repository, rpc noderpc.INode, network types.Network, address, key, protocol string, level int64) (int64, error) {
	symLink, err := bcd.GetProtoSymLink(protocol)
	if err != nil {
		return 0, err
	}
	storageTypeByte, err := contracts.ScriptPart(network, address, symLink, consts.STORAGE)
	if err != nil {
		return 0, err
	}
	storage, err := ast.NewTypedAstFromBytes(storageTypeByte)
	if err != nil {
		return 0, err
	}

	node := storage.FindByName(key, false)
	if node == nil {
		return 0, errors.Wrap(ErrBigMapNotFound, key)
	}

	storageJSON, err := rpc.GetScriptStorageRaw(address, level)
	if err != nil {
		return 0, err
	}
	var storageData ast.UntypedAST
	if err := json.Unmarshal(storageJSON, &storageData); err != nil {
		return 0, err
	}
	if err := storage.Settle(storageData); err != nil {
		return 0, err
	}

	if bm, ok := node.(*ast.BigMap); ok {
		return *bm.Ptr, nil
	}

	return 0, errors.Wrap(ErrBigMapNotFound, key)
}

// FindByName -
func FindByName(repo models.GeneralRepository, contracts contract.Repository, network types.Network, address, key, protocol string) *ast.BigMap {
	symLink, err := bcd.GetProtoSymLink(protocol)
	if err != nil {
		return nil
	}
	storageTypeByte, err := contracts.ScriptPart(network, address, symLink, consts.STORAGE)
	if err != nil {
		return nil
	}
	storage, err := ast.NewTypedAstFromBytes(storageTypeByte)
	if err != nil {
		return nil
	}
	node := storage.FindByName(key, false)
	if node == nil {
		return nil
	}
	if bm, ok := node.(*ast.BigMap); ok {
		return bm
	}

	return nil
}
