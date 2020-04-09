package metrics

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// SetBigMapDiffsStrings -
func (h *Handler) SetBigMapDiffsStrings(operationID string) error {
	arr, err := h.ES.GetBigMapDiffsJSONByOperationID(operationID)
	if err != nil {
		return err
	}
	if len(arr) == 0 {
		return nil
	}

	result := make([]models.BigMapDiff, len(arr))
	for i, bmd := range arr {
		var b models.BigMapDiff
		b.ParseElasticJSON(bmd)

		b.KeyStrings = stringer.Get(bmd.Get("_source.key"))
		value := gjson.Parse(bmd.Get("_source.value").String())
		b.ValueStrings = stringer.Get(value)
		result[i] = b
	}

	return h.ES.BulkUpdateBigMapDiffs(result)
}
