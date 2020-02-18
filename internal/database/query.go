package database

// GetUserByLogin -
func (d *db) GetUserByLogin(login string) *User {
	var user User
	d.ORM.Where("login = ?", login).First(&user)
	return &user
}

// CreateUser -
func (d *db) CreateUser(user User) error {
	return d.ORM.Create(&user).Error
}

// Close -
func (d *db) Close() {
	d.ORM.Close()
}
