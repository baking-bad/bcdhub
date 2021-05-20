package database

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/jinzhu/gorm"
)

// Account model
type Account struct {
	gorm.Model
	UserID        uint   `gorm:"primary_key;not null"`
	PrivateKey    string `gorm:"primary_key;not null"`
	PublicKeyHash string
	Network       types.Network
}

// GetOrCreateAccount -
func (d *db) GetOrCreateAccount(a *Account) error {
	return d.Where("user_id = ? AND private_key = ?", a.UserID, a.PrivateKey).FirstOrCreate(a).Error
}
