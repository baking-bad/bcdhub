package core

import "github.com/baking-bad/bcdhub/internal/models/types"

// DeleteByContract -
func (p *Postgres) DeleteByContract(network types.Network, indices []string, address string) error {
	for i := range indices {
		if _, err := p.DB.Model().Table(indices[i]).
			Where("network = ?", network).
			Where("contract = ?", address).
			Delete(); err != nil {
			return err
		}
	}
	return nil
}
