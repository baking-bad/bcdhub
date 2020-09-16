package database

import "github.com/jinzhu/gorm"

// Subscription model
type Subscription struct {
	gorm.Model
	UserID    uint   `gorm:"primary_key;not null"`
	Address   string `gorm:"primary_key;not null"`
	Network   string `gorm:"primary_key;not null"`
	Alias     string
	WatchMask uint
	SentryDSN string
}

func (d *db) GetSubscription(userID uint, address, network string) (s Subscription, err error) {
	err = d.
		Scopes(userIDScope(userID), networkScope(network), addressScope(address)).
		First(&s).Error
	return
}

func (d *db) GetSubscriptions(address, network string) ([]Subscription, error) {
	var subs []Subscription

	err := d.
		Scopes(contract(address, network)).
		Find(&subs).Error

	return subs, err
}

func (d *db) ListSubscriptions(userID uint) ([]Subscription, error) {
	var subs []Subscription

	err := d.
		Scopes(userIDScope(userID)).
		Order("created_at DESC").
		Find(&subs).Error

	return subs, err
}

func (d *db) UpsertSubscription(s *Subscription) error {
	return d.
		Scopes(userIDScope(s.UserID), contract(s.Address, s.Network)).
		Assign(Subscription{Alias: s.Alias, WatchMask: s.WatchMask, SentryDSN: s.SentryDSN}).
		FirstOrCreate(s).Error
}

func (d *db) DeleteSubscription(s *Subscription) error {
	return d.Unscoped().
		Scopes(userIDScope(s.UserID), contract(s.Address, s.Network)).
		Delete(Subscription{}).Error
}

func (d *db) GetSubscriptionsCount(address, network string) (count int, err error) {
	err = d.
		Model(&Subscription{}).
		Scopes(contract(address, network)).
		Count(&count).Error
	return
}
