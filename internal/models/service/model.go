package service

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// State -
type State struct {
	Name   string `gorm:"index:service_name_idx;unique;"`
	LastID int64
}

// GetID -
func (s *State) GetID() int64 {
	return 0
}

// GetIndex -
func (s *State) GetIndex() string {
	return "service_states"
}

// Save -
func (s *State) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "name"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_id": s.LastID,
		}),
	}).Create(s).Error
}
