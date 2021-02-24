package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// errors -
var (
	ErrBigMapNotFound = errors.New("Big map is not found")
)

// GetBigMapPtr -
func GetBigMapPtr(rpc noderpc.INode, address, key, network, protocol, sharePath string, level int64) (int64, error) {
	data, err := fetch.Contract(address, network, protocol, sharePath)
	if err != nil {
		return 0, err
	}
	script := gjson.ParseBytes(data)
	storage, err := ast.NewTypedAstFromString(script.Get("#(prim==\"storage\").args").Raw)
	if err != nil {
		return 0, err
	}

	node := storage.FindByName(key)
	if node == nil {
		return 0, errors.Wrap(ErrBigMapNotFound, key)
	}

	storageJSON, err := rpc.GetScriptStorageJSON(address, level)
	if err != nil {
		return 0, err
	}
	var storageData ast.UntypedAST
	if err := json.UnmarshalFromString(storageJSON.Raw, &storageData); err != nil {
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
