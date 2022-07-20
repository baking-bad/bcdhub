package contract

import (
	"math/rand"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
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
func (storage *Storage) Get(address string) (response contract.Contract, err error) {
	var accountID int64
	if err = storage.DB.Model((*account.Account)(nil)).Column("id").Where("address = ?", address).Select(&accountID); err != nil {
		return
	}

	err = storage.DB.Model(&response).Where("contract.account_id = ?", accountID).Relation("Account").Relation("Manager").Relation("Delegate").Relation("Alpha").Relation("Babylon").Relation("Jakarta").Select()
	return
}

// GetAll -
func (storage *Storage) GetAll(filters map[string]interface{}) (response []contract.Contract, err error) {
	query := storage.DB.Model((*contract.Contract)(nil))
	for key, value := range filters {
		query.Where("? = ?", pg.Ident(key), value)
	}
	err = query.Relation("Account").Relation("Manager").Relation("Delegate").Select(&response)
	return
}

// GetRandom -
func (storage *Storage) GetRandom() (response contract.Contract, err error) {
	var id int64
	if err = storage.DB.Model(&response).ColumnExpr("max(contract.id)").Select(&id); err != nil {
		return
	}

	err = storage.DB.Model(&response).Where("contract.id = ?", rand.Int63n(id)).
		Relation("Account").Relation("Manager").Relation("Delegate").Relation("Alpha").Relation("Babylon").Relation("Jakarta").First()
	return
}

// GetTokens -
func (storage *Storage) GetTokens(tokenInterface string, offset, size int64) ([]contract.Contract, int64, error) {
	tags := types.FA12Tag | types.FA1Tag | types.FA2Tag
	if tokenInterface == "fa1-2" || tokenInterface == "fa1" || tokenInterface == "fa2" {
		tags = types.NewTags([]string{tokenInterface})
	}

	query := storage.DB.Model((*contract.Contract)(nil)).
		Where("(tags & ?) > 0", tags).
		Order("id desc").
		Limit(storage.GetPageSize(size)).
		Offset(int(offset))

	var contracts []contract.Contract
	err := storage.DB.Model().TableExpr("(?) as contract", query).
		ColumnExpr("contract.*").
		ColumnExpr("account.address as account__address, account.alias as account__alias").
		ColumnExpr("manager.address as manager__address, manager.alias as manager__alias").
		ColumnExpr("delegate.address as delegate__address, delegate.alias as delegate__alias").
		Join(`LEFT JOIN "accounts" AS "account" ON "account"."id" = "contract"."account_id"`).
		Join(`LEFT JOIN "accounts" AS "manager" ON "manager"."id" = "contract"."manager_id" `).
		Join(`LEFT JOIN "accounts" AS "delegate" ON "delegate"."id" = "contract"."delegate_id"`).
		Select(&contracts)
	if err != nil {
		return nil, 0, err
	}

	count, err := storage.DB.Model((*contract.Contract)(nil)).Where("(contract.tags & ?) > 0", tags).Count()
	return contracts, int64(count), err
}

// SameCount -
func (storage *Storage) SameCount(c contract.Contract) (int, error) {
	query := storage.DB.Model((*contract.Contract)(nil))
	if c.AlphaID > 0 {
		query.WhereOr("alpha_id = ?", c.AlphaID)
	}
	if c.BabylonID > 0 {
		query.WhereOr("babylon_id = ?", c.BabylonID)
	}
	if c.JakartaID > 0 {
		query.WhereOr("jakarta_id = ?", c.JakartaID)
	}
	return query.Count()
}

// ByHash -
func (storage *Storage) ByHash(hash string) (result contract.Script, err error) {
	err = storage.DB.Model(&result).Where("hash = ?", hash).First()
	return
}

// Script -
func (storage *Storage) Script(address string, symLink string) (contract.Script, error) {
	var accountID int64
	if err := storage.DB.Model((*account.Account)(nil)).Column("id").Where("address = ?", address).Select(&accountID); err != nil {
		return contract.Script{}, err
	}

	var c contract.Contract
	query := storage.DB.Model(&c).Where("account_id = ?", accountID)
	switch symLink {
	case bcd.SymLinkAlpha:
		err := query.Relation("Alpha").Select()
		return c.Alpha, err
	case bcd.SymLinkBabylon:
		err := query.Relation("Babylon").Select()
		return c.Babylon, err
	case bcd.SymLinkJakarta:
		err := query.Relation("Jakarta").Select()
		return c.Jakarta, err
	}
	return c.Alpha, errors.Errorf("unknown protocol symbolic link: %s", symLink)
}

// GetScripts -
func (storage *Storage) GetScripts(limit, offset int) (scripts []contract.Script, err error) {
	err = storage.DB.Model(&scripts).
		ColumnExpr("tags, hash, project_id, fail_strings, annotations, entrypoints").
		Limit(limit).Offset(offset).Order("id asc").Select()
	return
}

// UpdateProjectID -
func (storage *Storage) UpdateProjectID(scripts []contract.Script) error {
	_, err := storage.DB.Model(&scripts).Set("project_id = _data.project_id").WherePK().Update()
	return err
}

// Code -
func (storage *Storage) Code(id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.Model((*contract.Script)(nil)).Where("id = ?", id).Column("code").Select(&data)
	return data, err
}

// Parameter -
func (storage *Storage) Parameter(id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.Model((*contract.Script)(nil)).Where("id = ?", id).Column("parameter").Select(&data)
	return data, err
}

// Storage -
func (storage *Storage) Storage(id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.Model((*contract.Script)(nil)).Where("id = ?", id).Column("storage").Select(&data)
	return data, err
}

// Storage -
func (storage *Storage) Views(id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.Model((*contract.Script)(nil)).Where("id = ?", id).Column("views").Select(&data)
	return data, err
}

// ScriptPart -
func (storage *Storage) ScriptPart(address string, symLink, part string) ([]byte, error) {
	var accountID int64
	if err := storage.DB.Model((*account.Account)(nil)).Column("id").Where("address = ?", address).Select(&accountID); err != nil {
		return nil, err
	}

	query := storage.DB.Model((*contract.Contract)(nil)).Where("account_id = ?", accountID)

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
	err := query.Select(pg.Scan(&data))
	return data, err
}

// RecentlyCalled -
func (storage *Storage) RecentlyCalled(offset, size int64) (contracts []contract.Contract, err error) {
	query := storage.DB.Model((*contract.Contract)(nil)).
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
		Select(&contracts)
	return
}
