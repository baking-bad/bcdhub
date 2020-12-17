package metrics

import "github.com/baking-bad/bcdhub/internal/models/contract"

// Manager -
type Manager struct{}

// NewManager -
func NewManager() *Manager {
	return &Manager{}
}

// Compute -
func (m *Manager) Compute(a, b contract.Contract) Feature {
	f := Feature{
		Name: "manager",
	}

	if a.Address == b.Address && a.Network == b.Network {
		f.Value = 1
	}
	return f
}
