package tokens

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
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
func (m Metadata) ToModel(address, network string) *models.TZIP {
	return &models.TZIP{
		Address: address,
		Network: network,
		Level:   m.Level,
		TZIP12: tzip.TZIP12{
			Tokens: &tzip.TokenMetadataType{
				Static: []tzip.TokenMetadata{
					{
						Symbol:          m.Symbol,
						Name:            m.Name,
						Decimals:        m.Decimals,
						TokenID:         m.TokenID,
						Extras:          m.Extras,
						RegistryAddress: m.RegistryAddress,
					},
				},
			},
		},
	}
}
