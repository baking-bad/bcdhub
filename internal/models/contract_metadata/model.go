package contract_metadata

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// ContractMetadata -
type ContractMetadata struct {
	// nolint
	tableName struct{} `pg:"contract_metadata"`

	ID         int64
	UpdatedAt  uint64 `pg:",use_zero"`
	Level      int64  `pg:",use_zero"`
	Timestamp  time.Time
	Address    string
	Network    types.Network `pg:",type:SMALLINT"`
	Slug       string
	DomainName string
	OffChain   bool                   `pg:",use_zero"`
	Extras     map[string]interface{} `pg:",type:jsonb"`

	TZIP16
	TZIP20
}

// BeforeInsert -
func (t *ContractMetadata) BeforeInsert(ctx context.Context) error {
	t.UpdatedAt = uint64(time.Now().Unix())
	return nil
}

// BeforeUpdate -
func (t *ContractMetadata) BeforeUpdate(ctx context.Context) (context.Context, error) {
	t.UpdatedAt = uint64(time.Now().Unix())
	return ctx, nil
}

// GetID -
func (t *ContractMetadata) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *ContractMetadata) GetIndex() string {
	return "contract_metadata"
}

// Save -
func (t *ContractMetadata) Save(tx pg.DBI) error {
	_, err := tx.Model(t).OnConflict("(id) DO UPDATE").
		Set(`
		updated_at = ?,
		level = excluded.level,
		timestamp = excluded.timestamp,
		extras = excluded.extras,
		events = excluded.events,
		name = excluded.name,
		description = excluded.description,
		version = excluded.version,
		license = excluded.license,
		homepage = excluded.homepage,
		authors = excluded.authors,
		interfaces = excluded.interfaces,
		slug = excluded.slug,
		views = excluded.views`, uint64(time.Now().Unix())).
		Returning("id").Insert()
	return err
}

// LogFields -
func (t *ContractMetadata) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network": t.Network,
		"address": t.Address,
		"level":   t.Level,
	}
}
