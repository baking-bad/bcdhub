package database

import (
	"database/sql/driver"

	"github.com/jinzhu/gorm"
)

// User model
type User struct {
	gorm.Model
	Login         string `gorm:"primary_key"`
	Name          string
	AvatarURL     string
	Token         string
	Subscriptions []Subscription
}

// Subscription model
type Subscription struct {
	gorm.Model
	UserID     uint       `gorm:"primary_key"`
	EntityID   string     `gorm:"primary_key"`
	EntityType EntityType `gorm:"primary_key" sql:"type:entity_type"`
}

// EntityType -
type EntityType string

const (
	project  EntityType = "project"
	contract EntityType = "contract"
)

// Scan -
func (e *EntityType) Scan(value interface{}) error {
	*e = EntityType(value.([]byte))
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
