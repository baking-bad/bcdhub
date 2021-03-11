package metrics

import (
	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// SetBigMapDiffsStrings -
func (h *Handler) SetBigMapDiffsStrings(bmd *bigmapdiff.BigMapDiff) error {
	keyStrings, err := getStrings(bmd.KeyBytes())
	if err != nil {
		return err
	}
	bmd.KeyStrings = keyStrings

	if bmd.Value != nil {
		valStrings, err := getStrings(bmd.ValueBytes())
		if err != nil {
			return err
		}
		bmd.ValueStrings = valStrings
	}
	return nil
}

func getStrings(data []byte) ([]string, error) {
	var tree ast.UntypedAST
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}
	return tree.GetStrings(true)
}
