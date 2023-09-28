package core

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	bcdLogger "github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Postgres -
type Postgres struct {
	DB   *bun.DB
	conn *sql.DB

	PageSize int64

	schema string
}

func connectionOptions(cfg Config, schema string, appName string) ([]pgdriver.Option, error) {
	opts := make([]pgdriver.Option, 0)

	if cfg.DBName != "" {
		opts = append(opts, pgdriver.WithDatabase(cfg.DBName))
	} else {
		return nil, errors.New("empty database name")
	}

	if cfg.Host != "" && cfg.Port > 0 {
		opts = append(opts, pgdriver.WithAddr(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)))
	} else {
		return nil, errors.Errorf("empty host or zero port: host=%s port=%d", cfg.Host, cfg.Port)
	}

	if cfg.User != "" {
		opts = append(opts, pgdriver.WithUser(cfg.User))
	} else {
		return nil, errors.New("empty database user")
	}

	if cfg.Password != "" {
		opts = append(opts, pgdriver.WithPassword(cfg.Password))
	} else {
		return nil, errors.New("empty database password")
	}

	if appName != "" {
		opts = append(opts, pgdriver.WithApplicationName(appName))
	}

	if cfg.SslMode != "" {
		switch cfg.SslMode {
		case "disable":
			opts = append(opts, pgdriver.WithInsecure(true))
		default:
		}
	}

	if schema != "" {
		opts = append(opts, pgdriver.WithConnParams(map[string]interface{}{
			"search_path": schema,
		}))
	}

	return opts, nil
}

// New -
func New(cfg Config, schemaName, appName string, opts ...PostgresOption) (*Postgres, error) {
	postgres := Postgres{
		schema: schemaName,
	}

	connectionOptions, err := connectionOptions(cfg, schemaName, appName)
	if err != nil {
		return nil, err
	}

	pgconn := pgdriver.NewConnector(connectionOptions...)
	postgres.conn = sql.OpenDB(pgconn)
	postgres.DB = bun.NewDB(postgres.conn, pgdialect.New())

	maxOpenConns := 4 * runtime.GOMAXPROCS(0)
	postgres.conn.SetMaxOpenConns(maxOpenConns)
	postgres.conn.SetMaxIdleConns(maxOpenConns)

	for _, opt := range opts {
		opt(&postgres)
	}

	return &postgres, nil
}

const (
	waitingTimeout = 10
)

// WaitNew - waiting for db up and creating connection
func WaitNew(cfg Config, schemaName, appName string, timeout int, opts ...PostgresOption) *Postgres {
	var db *Postgres
	var err error

	if timeout < 1 {
		timeout = waitingTimeout
	}

	for db == nil {
		db, err = New(cfg, schemaName, appName, opts...)
		if err != nil {
			bcdLogger.Warning().Msgf("Waiting postgres up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}

	for err := db.DB.Ping(); err != nil; err = db.DB.Ping() {
		bcdLogger.Warning().Msgf("Waiting postgres up %d seconds...", timeout)
		time.Sleep(time.Second * time.Duration(timeout))
	}

	// register many-to-many relationships
	db.DB.RegisterModel(models.ManyToMany()...)

	return db
}

func (p *Postgres) InitDatabase(ctx context.Context) error {
	if err := createSchema(ctx, p.DB, p.schema); err != nil {
		return err
	}

	if err := createTables(ctx, p.DB); err != nil {
		return err
	}

	if err := createBaseIndices(ctx, p.DB); err != nil {
		return err
	}

	return nil
}

// Close -
func (p *Postgres) Close() error {
	if err := p.conn.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}

// IsRecordNotFound -
func (p *Postgres) IsRecordNotFound(err error) bool {
	return err != nil && errors.Is(err, sql.ErrNoRows)
}

// Execute -
func (p *Postgres) Execute(rawSQL string) error {
	_, err := p.DB.Exec(rawSQL)
	return err
}
