package core

import (
	"errors"
	"fmt"
	"time"

	bcdLogger "github.com/baking-bad/bcdhub/internal/logger"
	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

// Postgres -
type Postgres struct {
	DB *gorm.DB

	PageSize int64
}

// New -
func New(connection, appName string, opts ...PostgresOption) (*Postgres, error) {
	pg := Postgres{}
	if appName != "" {
		connection = fmt.Sprintf("%s application_name=%s", connection, appName)
	}

	db, err := gorm.Open(postgres.Open(connection), &gorm.Config{
		Logger: newLogger(),
	})
	if err != nil {
		return nil, err
	}

	pg.DB = db

	for _, opt := range opts {
		opt(&pg)
	}

	sql, err := pg.DB.DB()
	if err != nil {
		return nil, err
	}

	sql.SetMaxOpenConns(200)
	sql.SetMaxIdleConns(100)

	return &pg, nil
}

const (
	waitingTimeout = 10
)

// WaitNew - waiting for db up and creating connection
func WaitNew(connectionString, appName string, timeout int, opts ...PostgresOption) *Postgres {
	var db *Postgres
	var err error

	if timeout < 1 {
		timeout = waitingTimeout
	}

	for db == nil {
		db, err = New(connectionString, appName, opts...)
		if err != nil {
			bcdLogger.Warning().Msgf("Waiting postgres up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}

	return db
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
	return err != nil && errors.Is(err, gorm.ErrRecordNotFound)
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

// Execute -
func (p *Postgres) Execute(rawSQL string) error {
	return p.DB.Exec(rawSQL).Error
}
