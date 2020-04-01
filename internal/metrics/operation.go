package metrics

import "github.com/baking-bad/bcdhub/internal/models"

// SetOperationAliases -
func (h *Handler) SetOperationAliases(op *models.Operation, aliases map[string]string) {
	op.SourceAlias = aliases[op.Source]
	op.DestinationAlias = aliases[op.Destination]
}
