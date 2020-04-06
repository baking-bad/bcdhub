package metrics

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/models"
)

// SetBigMapDiffsKeyString -
func (h *Handler) SetBigMapDiffsKeyString(operationID string) error {
	arr, err := h.ES.GetBigMapDiffsJSONByOperationID(operationID)
	if err != nil {
		return err
	}
	if len(arr) == 0 {
		return nil
	}

	result := make([]models.BigMapDiff, len(arr))
	for i, bmd := range arr {
		data := stringer.Get(bmd.Get("_source.key"))
		var b models.BigMapDiff
		b.ParseElasticJSON(bmd)
		b.KeyStrings = data
		result[i] = b
	}

	return h.ES.BulkUpdateBigMapDiffs(result)
}
