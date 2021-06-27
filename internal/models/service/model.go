package service

import "gorm.io/gorm"

// State -
type State struct {
	ID     int64
	Name   string `gorm:"index:service_name_idx"`
	LastID int64
}

// GetID -
func (s *State) GetID() int64 {
	return s.ID
}

// GetIndex -
func (s *State) GetIndex() string {
	return "service_states"
}

// Save -
func (s *State) Save(tx *gorm.DB) error {
	return tx.Save(s).Error
}
