package core

import (
	"fmt"
	"strings"
	"time"

	bcdLogger "github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	pg "github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

// Postgres -
type Postgres struct {
	DB *pg.DB

	PageSize int64
}

func parseConncetionString(connection string) (*pg.Options, error) {
	if len(connection) == 0 {
		return nil, errors.New("invalid connection string")
	}

	items := strings.Split(connection, " ")
	if len(items) == 0 {
		return nil, errors.Errorf("invalid connection string: %s", connection)
	}

	opts := new(pg.Options)
	var host string
	var port string
	for i := range items {
		values := strings.Split(items[i], "=")
		if len(values) != 2 {
			return nil, errors.Errorf("invalid connection string: %s", connection)
		}

		switch values[0] {
		case "host":
			host = values[1]
		case "user":
			opts.User = values[1]
		case "password":
			opts.Password = values[1]
		case "port":
			port = values[1]
		case "dbname":
			opts.Database = values[1]
		case "sslmode":
			// if values[1] != "disable" {
			// 	// TODO: init ssl mode
			// }
		}
	}

	opts.Addr = fmt.Sprintf("%s:%s", host, port)

	return opts, nil
}

// New -
func New(connection, appName string, opts ...PostgresOption) (*Postgres, error) {
	postgres := Postgres{}
	if appName != "" {
		connection = fmt.Sprintf("%s application_name=%s", connection, appName)
	}

	opt, err := parseConncetionString(connection)
	if err != nil {
		panic(err)
	}

	postgres.DB = pg.Connect(opt)

	for _, opt := range opts {
		opt(&postgres)
	}

	for _, model := range models.ManyToMany() {
		orm.RegisterTable(model)
	}

	return &postgres, nil
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
	return p.DB.Close()
}

// IsRecordNotFound -
func (p *Postgres) IsRecordNotFound(err error) bool {
	return err != nil && errors.Is(err, pg.ErrNoRows)
}

// OrStringArray
func OrStringArray(db *orm.Query, arr []string, fieldName string) *orm.Query {
	if len(arr) == 0 {
		return nil
	}

	str := fmt.Sprintf("%s = ?", fieldName)
	subQuery := db.Where(str, arr[0])
	for i := 1; i < len(arr); i++ {
		subQuery.WhereOr(str, arr[i])
	}
	return subQuery
}

// Execute -
func (p *Postgres) Execute(rawSQL string) error {
	_, err := p.DB.Exec(rawSQL)
	return err
}
