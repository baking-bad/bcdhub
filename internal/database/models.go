package database

import "github.com/jinzhu/gorm"

// User model
type User struct {
	gorm.Model
	Token      string
	Login      string
	Name       string
	AvatarURL  string
	Ownerships []*Ownership
}

// Ownership model
type Ownership struct {
	gorm.Model
	Address string
	Alias   string
	Type    string
	Users   []*User
}
