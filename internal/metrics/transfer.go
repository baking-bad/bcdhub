package metrics

import (
	"github.com/baking-bad/bcdhub/internal/models"
)

// SetTransferAliases -
func (h *Handler) SetTransferAliases(aliases map[string]string, transfer *models.Transfer) (bool, error) {
	var changed bool

	if alias, ok := aliases[transfer.From]; ok {
		transfer.FromAlias = alias
		changed = true
	}

	if alias, ok := aliases[transfer.To]; ok {
		transfer.ToAlias = alias
		changed = true
	}

	if alias, ok := aliases[transfer.Contract]; ok {
		transfer.Alias = alias
		changed = true
	}

	if alias, ok := aliases[transfer.Initiator]; ok {
		transfer.InitiatorAlias = alias
		changed = true
	}

	return changed, nil
}
