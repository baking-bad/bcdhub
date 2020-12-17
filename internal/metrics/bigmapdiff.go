package metrics

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
)

// SetBigMapDiffsStrings -
func (h *Handler) SetBigMapDiffsStrings(bmd *bigmapdiff.BigMapDiff) error {
	keyBytes, err := json.Marshal(bmd.Key)
	if err != nil {
		return err
	}
	bmd.KeyStrings = stringer.Get(string(keyBytes))
	bmd.ValueStrings = stringer.Get(bmd.Value)
	return nil
}
