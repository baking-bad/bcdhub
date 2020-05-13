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
}

// SubRating -
type SubRating struct {
	Count int `json:"count"`
	Users []struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatarURL"`
	} `json:"users"`
}

func (d *db) GetSubscription(address, network string) (s Subscription, err error) {
	err = d.ORM.Where("address = ? AND network = ?", address, network).Find(&s).Error
	return
}

func (d *db) ListSubscriptions(userID uint) ([]Subscription, error) {
	var subs []Subscription

	if err := d.ORM.Where("user_id = ?", userID).Order("created_at DESC").Find(&subs).Error; err != nil {
		return nil, err
	}

	return subs, nil
}

func (d *db) ListSubscriptionsWithLimit(userID uint, limit int) ([]Subscription, error) {
	var subs []Subscription

	if err := d.ORM.Order("created_at desc").Limit(limit).Where("user_id = ?", userID).Find(&subs).Error; err != nil {
		return nil, err
	}

	return subs, nil
}

func (d *db) UpsertSubscription(s *Subscription) error {
	return d.ORM.Where("network = ? AND address = ?", s.Network, s.Address).Assign(Subscription{WatchMask: s.WatchMask}).FirstOrCreate(s).Error
}

func (d *db) DeleteSubscription(s *Subscription) error {
	return d.ORM.Unscoped().Where("user_id = ? AND address = ? AND network = ?", s.UserID, s.Address, s.Network).Delete(Subscription{}).Error
}

func (d *db) GetSubscriptionRating(address, network string) (SubRating, error) {
	var s SubRating
	if err := d.ORM.Model(&Subscription{}).Where("address = ? AND network = ?", address, network).Count(&s.Count).Error; err != nil {
		return s, err
	}

	if err := d.ORM.Raw(`
		SELECT users.login, users.avatar_url
		FROM subscriptions
		INNER JOIN users ON subscriptions.user_id=users.id
		WHERE address = ? AND network = ? AND subscriptions.deleted_at IS NULL
		LIMIT 5;`, address, network).Scan(&s.Users).Error; err != nil {
		return s, err
	}

	return s, nil
}
