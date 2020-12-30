package metrics

import "github.com/baking-bad/bcdhub/internal/models/transfer"

// SetTransferAliases -
func (h *Handler) SetTransferAliases(transfer *transfer.Transfer) (bool, error) {
	var changed bool

	aliases, err := h.TZIP.GetAliasesMap(transfer.Network)
	if err != nil {
		if h.Storage.IsRecordNotFound(err) {
			err = nil
		}
		return changed, err
	}

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
