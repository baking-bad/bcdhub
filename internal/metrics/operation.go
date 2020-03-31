package metrics

import "github.com/baking-bad/bcdhub/internal/models"

// SetOperationAliases -
func (h *Handler) SetOperationAliases(op *models.Operation) error {
	aliases, err := h.DB.GetOperationAliases(op.Source, op.Destination, op.Network)
	if err != nil {
		return err
	}

	op.SourceAlias = aliases.Source
	op.DestinationAlias = aliases.Destination

	return nil
}
