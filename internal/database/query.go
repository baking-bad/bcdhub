package database

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

func (d *db) ListSubscriptions(userID uint) ([]Subscription, error) {
	var subs []Subscription

	if err := d.ORM.Where("user_id = ?", userID).Find(&subs).Error; err != nil {
		return nil, err
	}

	return subs, nil
}

func (d *db) CreateSubscription(s *Subscription) error {
	return d.ORM.Create(s).Error
}

func (d *db) DeleteSubscription(s *Subscription) error {
	return d.ORM.Delete(s).Error
}

func (d *db) Close() {
	d.ORM.Close()
}
