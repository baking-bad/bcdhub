package core

import (
	"context"
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

	schema string
}

func parseConnectionString(connection, schemaName string) (*pg.Options, error) {
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
		}
	}

	opts.Addr = fmt.Sprintf("%s:%s", host, port)
	opts.IdleTimeout = time.Second * 15
	opts.IdleCheckFrequency = time.Second * 10
	opts.OnConnect = func(ctx context.Context, cn *pg.Conn) error {
		schema := pg.Ident(schemaName)
		if _, err := cn.Exec("create schema if not exists ?", schema); err != nil {
			return err
		}
		_, err := cn.Exec("set search_path = ?", schema)
		return err
	}

	return opts, nil
}

// New -
func New(connection, schemaName, appName string, opts ...PostgresOption) (*Postgres, error) {
	postgres := Postgres{
		schema: schemaName,
	}
	if appName != "" {
		connection = fmt.Sprintf("%s application_name=%s", connection, appName)
	}

	opt, err := parseConnectionString(connection, schemaName)
	if err != nil {
		return nil, err
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
func WaitNew(connectionString, schemaName, appName string, timeout int, opts ...PostgresOption) *Postgres {
	var db *Postgres
	var err error

	if timeout < 1 {
		timeout = waitingTimeout
	}

	for db == nil {
		db, err = New(connectionString, schemaName, appName, opts...)
		if err != nil {
			bcdLogger.Warning().Msgf("Waiting postgres up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}

	for err := db.DB.Ping(context.Background()); err != nil; err = db.DB.Ping(context.Background()) {
		bcdLogger.Warning().Msgf("Waiting postgres up %d seconds...", timeout)
		time.Sleep(time.Second * time.Duration(timeout))
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

// Execute -
func (p *Postgres) Execute(rawSQL string) error {
	_, err := p.DB.Exec(rawSQL)
	return err
}
