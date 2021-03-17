package core

// DeleteByLevelAndNetwork -
func (p *Postgres) DeleteByLevelAndNetwork(indices []string, network string, maxLevel int64) error {
	for i := range indices {
		if err := p.DB.Unscoped().Table(indices[i]).
			Where("network = ?", network).
			Where("level > ?", maxLevel).
			Delete(nil).
			Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteByContract -
func (p *Postgres) DeleteByContract(indices []string, network, address string) error {
	for i := range indices {
		if err := p.DB.Unscoped().Table(indices[i]).
			Where("network = ?", network).
			Where("contract = ?", address).
			Delete(nil).
			Error; err != nil {
			return err
		}
	}
	return nil
}
