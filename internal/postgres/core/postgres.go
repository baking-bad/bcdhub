package core

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Postgres -
type Postgres struct {
	DB   *bun.DB
	conn *sql.DB

	PageSize int64

	schema    string
	timeout   time.Duration
	hasLogger bool
}

func buildDSN(cfg Config) (string, error) {
	if cfg.Host == "" || cfg.Port == 0 {
		return "", errors.Errorf("empty host or zero port: host=%s port=%d", cfg.Host, cfg.Port)
	}

	if cfg.DBName == "" {
		return "", errors.New("empty database name")
	}

	if cfg.User == "" {
		return "", errors.New("empty database user")
	}

	if cfg.Password == "" {
		return "", errors.New("empty database password")
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.DBName,
	}

	q := u.Query()
	if cfg.SslMode != "" {
		q.Set("sslmode", cfg.SslMode)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// New -
func New(cfg Config, schemaName, appName string, opts ...PostgresOption) (*Postgres, error) {
	postgres := Postgres{
		schema: schemaName,
	}
	for _, opt := range opts {
		opt(&postgres)
	}

	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "buildDSN")
	}
	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pgxConfig.RuntimeParams["application_name"] = appName
	if schemaName != "" {
		pgxConfig.RuntimeParams["search_path"] = schemaName
	}
	pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pgxConfig.StatementCacheCapacity = 0
	pgxConfig.DescriptionCacheCapacity = 0
	if postgres.timeout == 0 {
		postgres.timeout = 60 * time.Second
	}

	pgxConfig.ConnectTimeout = postgres.timeout
	pgxConfig.RuntimeParams["statement_timeout"] = strconv.Itoa(int(postgres.timeout / time.Millisecond))

	postgres.conn = stdlib.OpenDB(*pgxConfig)
	postgres.DB = bun.NewDB(postgres.conn, pgdialect.New())

	maxOpenConns := 4 * runtime.GOMAXPROCS(0)
	postgres.conn.SetMaxOpenConns(maxOpenConns)
	postgres.conn.SetMaxIdleConns(maxOpenConns)

	if postgres.hasLogger {
		postgres.DB.AddQueryHook(&logQueryHook{})
	}

	// register many-to-many relationships
	postgres.DB.RegisterModel(models.ManyToMany()...)

	return &postgres, nil
}

const (
	waitingTimeout = 10
)

// WaitNew - waiting for db up and creating connection
func WaitNew(ctx context.Context, cfg Config, schemaName, appName string, timeout int, opts ...PostgresOption) *Postgres {
	var db *Postgres
	var err error

	if timeout < 1 {
		timeout = waitingTimeout
	}

	for db == nil {
		db, err = New(cfg, schemaName, appName, opts...)
		if err != nil {
			log.Warn().Msgf("Waiting postgres up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}

	for err := db.DB.PingContext(ctx); err != nil; err = db.DB.PingContext(ctx) {
		log.Warn().Msgf("Waiting postgres up %d seconds...", timeout)
		time.Sleep(time.Second * time.Duration(timeout))
	}

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
