package database

import "github.com/jinzhu/gorm"

// Account model
type Account struct {
	gorm.Model
	UserID        uint   `gorm:"primary_key;not null"`
	PrivateKey    string `gorm:"primary_key;not null"`
	PublicKeyHash string
	Network       string
}

// GetOrCreateAccount -
func (d *db) GetOrCreateAccount(a *Account) error {
	return d.ORM.Where("user_id = ? AND private_key = ?", a.UserID, a.PrivateKey).FirstOrCreate(a).Error
}
