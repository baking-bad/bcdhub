package database

import (
	"database/sql/driver"

	"github.com/jinzhu/gorm"
)

// User model
type User struct {
	gorm.Model
	Login         string `gorm:"primary_key;not null"`
	Name          string
	AvatarURL     string `gorm:"not null"`
	Token         string `gorm:"not null"`
	Subscriptions []Subscription
}

// Subscription model
type Subscription struct {
	gorm.Model
	UserID     uint       `gorm:"primary_key;not null"`
	EntityID   string     `gorm:"primary_key;not null"`
	EntityType EntityType `gorm:"primary_key;not null;type:varchar(8)" sql:"type:entity_type"`
}

// EntityType -
type EntityType string

// Scan -
func (e *EntityType) Scan(value interface{}) error {
	*e = EntityType(value.(string))
	return nil
}

// Value -
func (e EntityType) Value() (driver.Value, error) {
	return string(e), nil
}

// SubRating -
type SubRating struct {
	Count int `json:"count"`
	Users []struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatarURL"`
	} `json:"users"`
}

// Alias -
type Alias struct {
	ID      int64 `gorm:"primary_key,AUTO_INCREMENT"`
	Alias   string
	Network string
	Address string
}

// OperationAlises -
type OperationAlises struct {
	Source      string
	Destination string
}
