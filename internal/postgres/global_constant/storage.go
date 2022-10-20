package global_constant

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
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
func (storage *Storage) Get(address string) (response contract.GlobalConstant, err error) {
	query := storage.DB.Model(&response)
	core.Address(address)(query)
	err = query.First()
	return
}

// All -
func (storage *Storage) All(addresses ...string) (response []contract.GlobalConstant, err error) {
	if len(addresses) == 0 {
		return
	}

	err = storage.DB.Model((*contract.GlobalConstant)(nil)).Where("address IN (?)", pg.In(addresses)).Select(&response)
	return
}

// List -
func (storage *Storage) List(size, offset int64, orderBy, sort string) ([]contract.ListGlobalConstantItem, error) {
	if offset < 0 {
		return nil, nil
	}
	if size < 1 || size > consts.MaxSize {
		size = consts.DefaultSize
	}

	lowerSort := strings.ToLower(sort)
	if lowerSort != "asc" && lowerSort != "desc" {
		lowerSort = "desc"
	}

	switch orderBy {
	case "level", "timestamp", "address":
		orderBy = fmt.Sprintf("%s %s", orderBy, lowerSort)
	case "links_count":
		orderBy = fmt.Sprintf("links_count %s, timestamp %s", lowerSort, lowerSort)
	default:
		orderBy = fmt.Sprintf("links_count %s, timestamp %s", lowerSort, lowerSort)
	}

	var constants []contract.ListGlobalConstantItem
	_, err := storage.DB.Query(&constants,
		`select global_constants.timestamp, global_constants.level, global_constants.address, count(contracts.id) as links_count from global_constants 
		left join script_constants as t on  global_constants.id = t.global_constant_id
		left join contracts on t.script_id = contracts.babylon_id or t.script_id = contracts.jakarta_id 
		group by global_constants.id
		order by ?
		limit ?
		offset ?`, pg.Safe(orderBy), size, offset)
	if err != nil {
		return nil, err
	}
	if len(constants) == 0 {
		constants = make([]contract.ListGlobalConstantItem, 0)
	}

	return constants, nil

}

// ForContract -
func (storage *Storage) ForContract(address string, size, offset int64) (response []contract.GlobalConstant, err error) {
	if offset < 0 || address == "" {
		return nil, nil
	}
	if size < 1 || size > consts.MaxSize {
		size = consts.DefaultSize
	}

	err = storage.DB.Model((*contract.Contract)(nil)).
		ColumnExpr("global_constants.*").
		Join("LEFT JOIN accounts on account_id = accounts.id").
		Join("LEFT JOIN script_constants as t on t.script_id = jakarta_id or t.script_id = babylon_id").
		Join("LEFT JOIN global_constants on t.global_constant_id = global_constants.id").
		Where("accounts.address = ?", address).
		Where("global_constant_id is not null").
		Limit(int(size)).
		Offset(int(offset)).
		Order("id desc").
		Select(&response)
	return
}

// ContractList -
func (storage *Storage) ContractList(address string, size, offset int64) ([]contract.Contract, error) {
	if offset < 0 || address == "" {
		return nil, nil
	}
	if size < 1 || size > consts.MaxSize {
		size = consts.DefaultSize
	}

	var id uint64
	if err := storage.DB.Model((*contract.GlobalConstant)(nil)).
		Column("id").
		Where("address = ?", address).
		Select(&id); err != nil {
		return nil, err
	}

	var scriptIDs []uint64
	if err := storage.DB.Model(new(contract.ScriptConstants)).
		DistinctOn("script_id").
		Column("script_id").
		Where("global_constant_id = ?", id).
		Select(&scriptIDs); err != nil {
		return nil, err
	}
	if len(scriptIDs) == 0 {
		return []contract.Contract{}, nil
	}

	var contracts []contract.Contract
	if err := storage.DB.Model(&contracts).
		ColumnExpr("contract.*").
		ColumnExpr("accounts.address as account__address, accounts.alias as account__alias").
		WhereIn("contract.babylon_id IN (?)", scriptIDs).
		WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
			q.WhereIn("contract.jakarta_id IN (?)", scriptIDs)
			return q, nil
		}).
		Join("LEFT JOIN accounts on contract.account_id = accounts.id").
		Order("contract.id desc").
		Limit(int(size)).
		Offset(int(offset)).
		Select(&contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}
