package metrics

// SetBigMapDiffsKeyString -
func (h *Handler) SetBigMapDiffsKeyString(operationID string) error {
	bmd, err := h.ES.GetBigMapDiffsByOperationID(operationID)
	if err != nil {
		return err
	}

	// Update bmd here

	return h.ES.BulkUpdateBigMapDiffs(bmd)
}
