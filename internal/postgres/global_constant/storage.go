package global_constant

import (
	"context"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
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
func (storage *Storage) Get(ctx context.Context, address string) (response contract.GlobalConstant, err error) {
	query := storage.DB.NewSelect().Model(&response)
	query = core.Address(query, address)
	err = query.Limit(1).Scan(ctx)
	return
}

// All -
func (storage *Storage) All(ctx context.Context, addresses ...string) (response []contract.GlobalConstant, err error) {
	if len(addresses) == 0 {
		return
	}

	err = storage.DB.NewSelect().Model(&response).Where("address IN (?)", bun.In(addresses)).Scan(ctx)
	return
}

// List -
func (storage *Storage) List(ctx context.Context, size, offset int64, orderBy, sort string) ([]contract.ListGlobalConstantItem, error) {
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
	err := storage.DB.NewRaw(
		`select global_constants.timestamp, global_constants.level, global_constants.address, count(contracts.id) as links_count from global_constants 
		left join script_constants as t on  global_constants.id = t.global_constant_id
		left join contracts on t.script_id = contracts.babylon_id or t.script_id = contracts.jakarta_id 
		group by global_constants.id
		order by ?
		limit ?
		offset ?`, bun.Safe(orderBy), size, offset).Scan(ctx, &constants)
	if err != nil {
		return nil, err
	}
	if len(constants) == 0 {
		constants = make([]contract.ListGlobalConstantItem, 0)
	}

	return constants, nil

}

// ForContract -
func (storage *Storage) ForContract(ctx context.Context, address string, size, offset int64) (response []contract.GlobalConstant, err error) {
	if offset < 0 || address == "" {
		return nil, nil
	}
	if size < 1 || size > consts.MaxSize {
		size = consts.DefaultSize
	}

	var accountID int64
	if err = storage.DB.
		NewSelect().
		Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Scan(ctx, &accountID); err != nil {
		return
	}

	var contr contract.Contract
	if err = storage.DB.
		NewSelect().
		Column("id", "alpha_id", "babylon_id", "jakarta_id").
		Model(&contr).
		Column("id").
		Where("account_id = ?", accountID).
		Scan(ctx); err != nil {
		return
	}

	ids := make([]int64, 0)
	if contr.BabylonID > 0 {
		ids = append(ids, contr.BabylonID)
	}
	if contr.JakartaID > 0 {
		ids = append(ids, contr.JakartaID)
	}
	if len(ids) == 0 {
		return nil, nil
	}

	err = storage.DB.NewSelect().Model((*contract.ScriptConstants)(nil)).
		ColumnExpr("global_constants.*").
		Join("LEFT JOIN global_constants on global_constant_id = global_constants.id").
		Where("global_constant_id is not null").
		Where("script_id IN (?)", bun.In(ids)).
		Limit(int(size)).
		Offset(int(offset)).
		Order("global_constants.id desc").
		Scan(ctx, &response)
	return
}

// ContractList -
func (storage *Storage) ContractList(ctx context.Context, address string, size, offset int64) ([]contract.Contract, error) {
	if offset < 0 || address == "" {
		return nil, nil
	}
	if size < 1 || size > consts.MaxSize {
		size = consts.DefaultSize
	}

	var id uint64
	if err := storage.DB.NewSelect().Model((*contract.GlobalConstant)(nil)).
		Column("id").
		Where("address = ?", address).
		Scan(ctx, &id); err != nil {
		return nil, err
	}

	var scriptIDs []uint64
	if err := storage.DB.NewSelect().Model(new(contract.ScriptConstants)).
		DistinctOn("script_id").
		Column("script_id").
		Where("global_constant_id = ?", id).
		Scan(ctx, &scriptIDs); err != nil {
		return nil, err
	}
	if len(scriptIDs) == 0 {
		return []contract.Contract{}, nil
	}

	var contracts []contract.Contract
	if err := storage.DB.NewSelect().Model(&contracts).
		ColumnExpr("contract.*").
		ColumnExpr("accounts.address as account__address, accounts.alias as account__alias").
		Where("contract.babylon_id IN (?)", bun.In(scriptIDs)).
		WhereOr("contract.jakarta_id IN (?)", bun.In(scriptIDs)).
		Join("LEFT JOIN accounts on contract.account_id = accounts.id").
		Order("contract.id desc").
		Limit(int(size)).
		Offset(int(offset)).
		Scan(ctx); err != nil {
		return nil, err
	}
	return contracts, nil
}
