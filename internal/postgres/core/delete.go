package core

// DeleteByContract -
func (p *Postgres) DeleteByContract(indices []string, address string) error {
	for i := range indices {
		if _, err := p.DB.Model().Table(indices[i]).
			Where("contract = ?", address).
			Delete(); err != nil {
			return err
		}
	}
	return nil
}
