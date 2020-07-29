package database

import "github.com/jinzhu/gorm"

// Verification model
type Verification struct {
	gorm.Model
	UserID    uint   `gorm:"primary_key;not null"`
	Address   string `gorm:"primary_key;not null"`
	Network   string `gorm:"primary_key;not null"`
	State     string
	SourceURL string
}
