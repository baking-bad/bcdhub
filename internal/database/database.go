package database

import (
	"github.com/jinzhu/gorm"
	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// DB -
type DB interface {
	GetOrCreateUser(*User) error
	GetUser(uint) (*User, error)
	GetSubscription(string, string) (Subscription, error)
	ListSubscriptions(uint) ([]Subscription, error)
	ListSubscriptionsWithLimit(uint, int) ([]Subscription, error)
	CreateSubscription(*Subscription) error
	DeleteSubscription(*Subscription) error
	GetSubscriptionRating(string) (SubRating, error)
	Close()
}

type db struct {
	ORM *gorm.DB
}

// New - creates db connection
func New(connectionString string) (DB, error) {
	gormDB, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	gormDB.LogMode(false)

	if !gormDB.HasTable(&Subscription{}) {
		gormDB.Exec("CREATE TYPE entity_type AS ENUM ('unknown','project','contract');")
	}

	gormDB.AutoMigrate(&User{}, &Subscription{}, &Alias{})

	gormDB = gormDB.Set("gorm:auto_preload", false)

	return &db{
		ORM: gormDB,
	}, nil
}
