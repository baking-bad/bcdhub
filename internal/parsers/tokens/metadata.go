package tokens

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
)

// Metadata -
type Metadata struct {
	RegistryAddress string
	Level           int64
	Timestamp       time.Time
	TokenID         int64
	Symbol          string
	Name            string
	Decimals        int64
	Extras          map[string]interface{}
}

// IsEmpty -
func (m Metadata) IsEmpty() bool {
	return m.Decimals == 0 &&
		len(m.Extras) == 0 &&
		m.Symbol == "" &&
		m.Name == ""
}

// ToModel -
func (m Metadata) ToModel(address, network string) *models.TokenMetadata {
	return &models.TokenMetadata{
		ID:              helpers.GenerateID(),
		Contract:        address,
		RegistryAddress: m.RegistryAddress,
		Network:         network,
		Timestamp:       m.Timestamp,
		TokenID:         m.TokenID,
		Symbol:          m.Symbol,
		Name:            m.Name,
		Decimals:        m.Decimals,
		Level:           m.Level,
		Extras:          m.Extras,
	}
}
