package service

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/models/service"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"gorm.io/gorm"
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
func (s *Storage) Get(name string) (state service.State, err error) {
	err = s.DB.Where("name = ?", name).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		state.Name = name
	}
	return
}

// Save -
func (s *Storage) Save(state service.State) error {
	return state.Save(s.DB)
}
