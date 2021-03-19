package core

import (
	"errors"
	"fmt"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

// Postgres -
type Postgres struct {
	DB *gorm.DB
}

// NewPostgres -
func NewPostgres(connection string) (*Postgres, error) {
	db, err := gorm.Open(postgres.Open(connection), &gorm.Config{
		Logger: newLogger(),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return &Postgres{
		DB: db,
	}, nil
}

// Close -
func (p *Postgres) Close() error {
	db, err := p.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// IsRecordNotFound -
func (p *Postgres) IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// OrStringArray
func OrStringArray(db *gorm.DB, arr []string, fieldName string) *gorm.DB {
	if len(arr) == 0 {
		return nil
	}

	str := fmt.Sprintf("%s = ?", fieldName)
	subQuery := db.Where(str, arr[0])
	for i := 1; i < len(arr); i++ {
		subQuery.Or(str, arr[i])
	}
	return subQuery
}

// AddTrigger -
func (p *Postgres) AddTrigger(rawSQL string) error {
	return p.DB.Exec(rawSQL).Error
}
