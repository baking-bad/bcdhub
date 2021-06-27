package models

import (
	"gorm.io/gorm"
)

// Model -
type Model interface {
	GetID() int64
	GetIndex() string
	Save(tx *gorm.DB) error
}
