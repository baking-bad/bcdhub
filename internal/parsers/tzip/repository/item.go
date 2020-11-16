package repository

import (
	"encoding/json"
	"time"

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
	t, err := time.Parse("2006 01 02 15 04", "2018 06 30 00 00")
	if err != nil {
		return nil, err
	}
	model := models.TZIP{
		Network:   item.Network,
		Address:   item.Address,
		Timestamp: t.UTC(),
	}

	err = json.Unmarshal(item.Metadata, &model)
	return &model, err
}
