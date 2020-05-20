package database

import (
	"time"

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
	MarkReadAt    time.Time
}

func (d *db) GetOrCreateUser(u *User) error {
	return d.ORM.Where("login = ?", u.Login).FirstOrCreate(u).Error
}

func (d *db) GetUser(userID uint) (*User, error) {
	var user User

	if err := d.ORM.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
