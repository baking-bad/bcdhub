package metrics

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// SetOperationAliases -
func (h *Handler) SetOperationAliases(aliases map[string]string, op *models.Operation) bool {
	var changed bool

	if srcAlias, ok := aliases[op.Source]; ok {
		op.SourceAlias = srcAlias
		changed = true
	}

	if dstAlias, ok := aliases[op.Destination]; ok {
		op.DestinationAlias = dstAlias
		changed = true
	}

	if dlgtAlias, ok := aliases[op.Delegate]; ok {
		op.DelegateAlias = dlgtAlias
		changed = true
	}

	return changed
}

// SetOperationBurned -
func (h *Handler) SetOperationBurned(op *models.Operation) {
	if op.Status != consts.Applied {
		return
	}

	if op.Result == nil {
		return
	}

	var burned int64

	if op.Result.PaidStorageSizeDiff != 0 {
		burned += op.Result.PaidStorageSizeDiff * 1000
	}

	if op.Result.AllocatedDestinationContract {
		burned += 257000
	}

	op.Burned = burned
}

// SetOperationStrings -
func (h *Handler) SetOperationStrings(op *models.Operation) {
	params := gjson.Parse(op.Parameters)
	op.ParameterStrings = stringer.Get(params)
	storage := gjson.Parse(op.DeffatedStorage)
	op.StorageStrings = stringer.Get(storage)
}
