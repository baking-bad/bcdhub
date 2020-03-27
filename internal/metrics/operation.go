package metrics

import "github.com/baking-bad/bcdhub/internal/models"

// SetOperationAliases -
func (h *Handler) SetOperationAliases(op *models.Operation) error {
	aliasSource, err := h.DB.GetAlias(op.Source, op.Network)
	if err != nil {
		return err
	}

	op.SourceAlias = aliasSource

	aliasDest, err := h.DB.GetAlias(op.Destination, op.Network)
	if err != nil {
		return err
	}

	op.DestinationAlias = aliasDest

	return nil
}
