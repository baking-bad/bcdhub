package metrics

import (
	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// SetBigMapDiffsStrings -
func (h *Handler) SetBigMapDiffsStrings(bmd *bigmapdiff.BigMapDiff) error {
	keyStrings, err := getStrings(bmd.Key)
	if err != nil {
		return err
	}
	bmd.KeyStrings = keyStrings

	valStrings, err := getStrings(bmd.Value)
	if err != nil {
		return err
	}
	bmd.ValueStrings = valStrings
	return nil
}

func getStrings(data []byte) ([]string, error) {
	var tree ast.UntypedAST
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}
	return tree.GetStrings(true)
}
