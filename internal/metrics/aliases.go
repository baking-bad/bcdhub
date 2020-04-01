package metrics

// GetAliases -
func (h *Handler) GetAliases(network string) (map[string]string, error) {
	aliasesFromDB, err := h.DB.GetAliases(network)
	if err != nil {
		return nil, err
	}

	aliases := make(map[string]string, len(aliasesFromDB))

	for _, a := range aliasesFromDB {
		aliases[a.Address] = a.Alias
	}

	return aliases, nil
}
