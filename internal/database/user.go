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
	Provider      string
	Subscriptions []Subscription
	MarkReadAt    time.Time
}

func (d *db) GetOrCreateUser(u *User, token string) error {
	err := d.Scopes(loginScope(u.Login)).First(u).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return d.Create(u).Error
		}

		return err
	}

	return d.Model(u).Scopes(loginScope(u.Login)).Update("token", u.Token).Error
}

func (d *db) GetUser(userID uint) (*User, error) {
	var user User

	if err := d.Scopes(idScope(userID)).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *db) UpdateUserMarkReadAt(userID uint, ts int64) error {
	return d.Model(&User{}).Scopes(idScope(userID)).Update("mark_read_at", time.Unix(ts, 0)).Error
}
