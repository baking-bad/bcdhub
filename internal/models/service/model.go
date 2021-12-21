package service

import (
	"github.com/go-pg/pg/v10"
)

// State -
type State struct {
	// nolint
	tableName struct{} `pg:"states"`

	ID     int64
	Name   string `pg:",unique"`
	LastID int64
}

// GetID -
func (s *State) GetID() int64 {
	return s.ID
}

// GetIndex -
func (s *State) GetIndex() string {
	return "states"
}

// Save -
func (s *State) Save(tx pg.DBI) error {
	_, err := tx.Model(s).
		OnConflict("(name) DO UPDATE").
		Set("last_id = EXCLUDED.last_id").
		Returning("id").
		Insert()

	return err
}
