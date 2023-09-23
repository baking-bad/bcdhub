package contract

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(ctx context.Context, address string) (response contract.Contract, err error) {
	var accountID int64
	if err = storage.DB.NewSelect().
		Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Scan(ctx, &accountID); err != nil {
		return
	}

	err = storage.DB.NewSelect().
		Model(&response).
		Where("contract.account_id = ?", accountID).
		Relation("Account").Relation("Manager").
		Relation("Delegate").Relation("Alpha").
		Relation("Babylon").Relation("Jakarta").
		Scan(ctx)
	return
}

// GetAll -
func (storage *Storage) GetAll(ctx context.Context, filters map[string]interface{}) (response []contract.Contract, err error) {
	query := storage.DB.NewSelect().Model(response)
	for key, value := range filters {
		query.Where("? = ?", bun.Ident(key), value)
	}
	err = query.Relation("Account").Relation("Manager").Relation("Delegate").Scan(ctx)
	return
}

// GetTokens -
func (storage *Storage) GetTokens(ctx context.Context, tokenInterface string, offset, size int64) ([]contract.Contract, int64, error) {
	tags := types.FA12Tag | types.FA1Tag | types.FA2Tag
	if tokenInterface == "fa1-2" || tokenInterface == "fa1" || tokenInterface == "fa2" {
		tags = types.NewTags([]string{tokenInterface})
	}

	query := storage.DB.NewSelect().Model((*contract.Contract)(nil)).
		Where("(tags & ?) > 0", tags).
		Order("id desc").
		Limit(storage.GetPageSize(size)).
		Offset(int(offset))

	var contracts []contract.Contract
	err := storage.DB.NewSelect().TableExpr("(?) as contract", query).
		ColumnExpr("contract.*").
		ColumnExpr("account.address as account__address, account.alias as account__alias").
		ColumnExpr("manager.address as manager__address, manager.alias as manager__alias").
		ColumnExpr("delegate.address as delegate__address, delegate.alias as delegate__alias").
		Join(`LEFT JOIN "accounts" AS "account" ON "account"."id" = "contract"."account_id"`).
		Join(`LEFT JOIN "accounts" AS "manager" ON "manager"."id" = "contract"."manager_id" `).
		Join(`LEFT JOIN "accounts" AS "delegate" ON "delegate"."id" = "contract"."delegate_id"`).
		Scan(ctx, &contracts)
	if err != nil {
		return nil, 0, err
	}

	count, err := storage.DB.NewSelect().Model((*contract.Contract)(nil)).
		Where("(contract.tags & ?) > 0", tags).
		Count(ctx)
	return contracts, int64(count), err
}

// ByHash -
func (storage *Storage) ByHash(ctx context.Context, hash string) (result contract.Script, err error) {
	err = storage.DB.NewSelect().Model(&result).Where("hash = ?", hash).Limit(1).Scan(ctx)
	return
}

// Script -
func (storage *Storage) Script(ctx context.Context, address string, symLink string) (contract.Script, error) {
	var accountID int64
	if err := storage.DB.
		NewSelect().
		Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Scan(ctx, &accountID); err != nil {
		return contract.Script{}, err
	}

	var c contract.Contract
	query := storage.DB.NewSelect().Model(&c).Where("account_id = ?", accountID)
	switch symLink {
	case bcd.SymLinkAlpha:
		err := query.Relation("Alpha").Scan(ctx)
		return c.Alpha, err
	case bcd.SymLinkBabylon:
		err := query.Relation("Babylon").Scan(ctx)
		return c.Babylon, err
	case bcd.SymLinkJakarta:
		err := query.Relation("Jakarta").Scan(ctx)
		return c.Jakarta, err
	}
	return c.Alpha, errors.Errorf("unknown protocol symbolic link: %s", symLink)
}

// GetScripts -
func (storage *Storage) GetScripts(ctx context.Context, limit, offset int) (scripts []contract.Script, err error) {
	err = storage.DB.NewSelect().Model(&scripts).
		ColumnExpr("id, tags, hash, fail_strings, annotations, entrypoints").
		Limit(limit).Offset(offset).Order("id asc").Scan(ctx)
	return
}

// Code -
func (storage *Storage) Code(ctx context.Context, id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.NewSelect().Model((*contract.Script)(nil)).Where("id = ?", id).Column("code").Scan(ctx, &data)
	return data, err
}

// Parameter -
func (storage *Storage) Parameter(ctx context.Context, id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.NewSelect().
		Model((*contract.Script)(nil)).
		Where("id = ?", id).
		Column("parameter").
		Scan(ctx, &data)
	return data, err
}

// Storage -
func (storage *Storage) Storage(ctx context.Context, id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.NewSelect().
		Model((*contract.Script)(nil)).
		Where("id = ?", id).
		Column("storage").
		Scan(ctx, &data)
	return data, err
}

// Storage -
func (storage *Storage) Views(ctx context.Context, id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.NewSelect().
		Model((*contract.Script)(nil)).
		Where("id = ?", id).
		Column("views").
		Scan(ctx, &data)
	return data, err
}

// ScriptPart -
func (storage *Storage) ScriptPart(ctx context.Context, address string, symLink, part string) ([]byte, error) {
	var accountID int64
	if err := storage.DB.NewSelect().
		Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Scan(ctx, &accountID); err != nil {
		return nil, err
	}

	query := storage.DB.NewSelect().Model((*contract.Contract)(nil)).Where("account_id = ?", accountID)

	switch symLink {
	case bcd.SymLinkAlpha:
		switch part {
		case consts.PARAMETER:
			query.Column("alpha.parameter").Relation("Alpha._")
		case consts.CODE:
			query.Column("alpha.code").Relation("Alpha._")
		case consts.STORAGE:
			query.Column("alpha.storage").Relation("Alpha._")
		case consts.VIEWS:
			query.Column("alpha.views").Relation("Alpha._")
		default:
			return nil, errors.Errorf("unknown script part name: %s", part)
		}
	case bcd.SymLinkBabylon:
		switch part {
		case consts.PARAMETER:
			query.Column("babylon.parameter").Relation("Babylon._")
		case consts.CODE:
			query.Column("babylon.code").Relation("Babylon._")
		case consts.STORAGE:
			query.Column("babylon.storage").Relation("Babylon._")
		case consts.VIEWS:
			query.Column("babylon.views").Relation("Babylon._")
		default:
			return nil, errors.Errorf("unknown script part name: %s", part)
		}
	case bcd.SymLinkJakarta:
		switch part {
		case consts.PARAMETER:
			query.Column("jakarta.parameter").Relation("Jakarta._")
		case consts.CODE:
			query.Column("jakarta.code").Relation("Jakarta._")
		case consts.STORAGE:
			query.Column("jakarta.storage").Relation("Jakarta._")
		case consts.VIEWS:
			query.Column("jakarta.views").Relation("Jakarta._")
		default:
			return nil, errors.Errorf("unknown script part name: %s", part)
		}
	default:
		return nil, errors.Errorf("unknown protocol symbolic link: %s", symLink)
	}
	var data []byte
	err := query.Scan(ctx, &data)
	return data, err
}

// RecentlyCalled -
func (storage *Storage) RecentlyCalled(ctx context.Context, offset, size int64) (contracts []contract.Contract, err error) {
	query := storage.DB.NewSelect().Model(&contracts).
		ColumnExpr("contract.id, contract.tx_count, contract.last_action, contract.account_id").
		ColumnExpr("account.address as account__address, account.alias as account__alias").
		Join(`LEFT JOIN "accounts" AS "account" ON "account"."id" = "contract"."account_id"`)

	if offset > 0 {
		query.Offset(int(offset))
	}
	if size > 0 {
		query.Limit(int(size))
	} else {
		query.Limit(10)
	}
	err = query.
		OrderExpr("contract.last_action desc, contract.tx_count desc").
		Scan(ctx)
	return
}

// Count -
func (storage *Storage) Count(ctx context.Context) (int, error) {
	return storage.DB.NewSelect().Model((*contract.Contract)(nil)).Count(ctx)
}

// FindOne -
func (storage *Storage) FindOne(ctx context.Context, tags types.Tags) (result contract.Contract, err error) {
	err = storage.DB.NewSelect().Model(&result).
		Where("tags&? > 0", tags).
		ColumnExpr("contract.id, contract.tx_count, contract.last_action, contract.account_id, contract.timestamp, contract.level").
		ColumnExpr("account.address as account__address, account.alias as account__alias").
		Join(`LEFT JOIN "accounts" AS "account" ON "account"."id" = "contract"."account_id"`).
		Limit(1).
		Scan(ctx)
	return
}
