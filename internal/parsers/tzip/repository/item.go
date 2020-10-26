package repository

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/models"
)

// Item -
type Item struct {
	Network  string
	Address  string
	Metadata []byte
}

// ToModel -
func (item Item) ToModel() (*models.TZIP, error) {
	model := models.TZIP{
		Network: item.Network,
		Address: item.Address,
	}

	err := json.Unmarshal(item.Metadata, &model)
	return &model, err
}
