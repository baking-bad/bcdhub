package metrics

import "github.com/baking-bad/bcdhub/internal/models"

// SetOperationAliases -
func (h *Handler) SetOperationAliases(aliases map[string]string, op *models.Operation) {
	op.SourceAlias = aliases[op.Source]
	op.DestinationAlias = aliases[op.Destination]
}
