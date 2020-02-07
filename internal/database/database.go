package database

import (
	"github.com/jinzhu/gorm"
	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// New - creates db connection
func New(connectionString string) (*gorm.DB, error) {
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	db.AutoMigrate(&User{}, &Ownership{})

	db = db.Set("gorm:auto_preload", true)

	return db, nil
}
