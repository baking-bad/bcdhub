package core

import "github.com/baking-bad/bcdhub/internal/models/types"

// DeleteByContract -
func (p *Postgres) DeleteByContract(network types.Network, indices []string, address string) error {
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
