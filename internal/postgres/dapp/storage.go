package dapp

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(slug string) (d dapp.DApp, err error) {
	err = storage.DB.Model(&d).Where("slug = ?", slug).First()
	return
}

// Get -
func (storage *Storage) All() (d []dapp.DApp, err error) {
	err = storage.DB.Model().Table(models.DocDApps).Order("dapps.order asc").Select(&d)
	return
}
