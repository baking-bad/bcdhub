package metrics

import (
	"github.com/baking-bad/bcdhub/internal/models/operation"
)

// SetOperationStrings -
func (h *Handler) SetOperationStrings(op *operation.Operation) {
	ps, err := getStrings(op.Parameters)
	if err == nil {
		op.ParameterStrings = ps
	}
	ss, err := getStrings(op.DeffatedStorage)
	if err == nil {
		op.StorageStrings = ss
	}
}
