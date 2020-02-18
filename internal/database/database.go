package database

import (
	"github.com/jinzhu/gorm"
	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// DB -
type DB interface {
	GetUserByLogin(string) *User
	CreateUser(User) error
	Close()
}

type db struct {
	ORM *gorm.DB
}

// New - creates db connection
func New(connectionString string) (DB, error) {
	gorm, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	gorm.LogMode(true)

	gorm.AutoMigrate(&User{}, &Ownership{})

	gorm = gorm.Set("gorm:auto_preload", true)

	return &db{
		ORM: gorm,
	}, nil
}
