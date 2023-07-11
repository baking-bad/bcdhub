package operation

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/pkg/errors"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(es *core.Postgres) *Storage {
	return &Storage{es}
}

type opgForContract struct {
	Counter int64
	Hash    *string
	ID      int64
}

func (storage *Storage) getContractOPG(accountID int64, size uint64, filters map[string]interface{}) (response []opgForContract, err error) {
	subQuery := storage.DB.Model().Table(models.DocOperations).Column("hash", "counter", "id")

	if _, ok := filters["entrypoints"]; !ok {
		subQuery.Where("source_id = ? OR destination_id = ?", accountID, accountID)
	} else {
		subQuery.Where("destination_id = ?", accountID)
	}

	if err := prepareOperationFilters(subQuery, filters); err != nil {
		return nil, err
	}

	query := storage.DB.Model().TableExpr("(?) as foo", subQuery.Order("id desc").Limit(1000)).
		ColumnExpr("foo.hash, foo.counter, max(id) as id")

	limit := storage.GetPageSize(int64(size))
	query.GroupExpr("foo.hash, foo.counter").Order("id desc").Limit(limit)

	err = query.Select(&response)
	return
}

func prepareOperationFilters(query *orm.Query, filters map[string]interface{}) error {
	for k, v := range filters {
		if v != "" {
			switch k {
			case "from":
				query.Where("timestamp >= to_timestamp(?)", v)
			case "to":
				query.Where("timestamp <= to_timestamp(?)", v)
			case "entrypoints":
				query.WhereIn("entrypoint IN (?)", v)
			case "last_id":
				query.Where("id < ?", v)
			case "status":
				query.WhereIn("status IN (?)", v)
			default:
				return errors.Errorf("Unknown operation filter: %s %v", k, v)
			}
		}
	}
	return nil
}

// GetByContract -
func (storage *Storage) GetByAccount(acc account.Account, size uint64, filters map[string]interface{}) (po operation.Pageable, err error) {
	opg, err := storage.getContractOPG(acc.ID, size, filters)
	if err != nil {
		return
	}
	if len(opg) == 0 {
		return
	}

	query := storage.DB.Model((*operation.Operation)(nil)).WhereGroup(func(q *orm.Query) (*orm.Query, error) {
		for i := range opg {
			q.WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				if opg[i].Hash == nil {
					q.Where("operation.hash is null")
				} else {
					q.Where("operation.hash = ?", opg[i].Hash)
				}
				return q.Where("operation.counter = ?", opg[i].Counter), nil
			})
		}
		return q, nil
	}).Relation("Destination").Relation("Source").Relation("Initiator").Relation("Delegate")

	addOperationSorting(query)

	if err = query.Select(&po.Operations); err != nil {
		return
	}

	if len(po.Operations) == 0 {
		return
	}

	lastID := po.Operations[0].ID
	for _, op := range po.Operations[1:] {
		if op.ID > lastID {
			continue
		}
		lastID = op.ID
	}
	po.LastID = fmt.Sprintf("%d", lastID)
	return
}

// Last - get last operation by `filters` with not empty deffated_storage
func (storage *Storage) Last(filters map[string]interface{}, lastID int64) (operation.Operation, error) {
	var (
		current = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	)

	for current.Year() >= 2018 {
		query := storage.DB.Model((*operation.Operation)(nil)).
			Where("deffated_storage is not null").
			Where("timestamp >= ?", current).
			Where("timestamp < ?", current.AddDate(1, 0, 0)).
			OrderExpr("operation.id desc")

		for key, value := range filters {
			query.Where("? = ?", pg.Ident(key), value)
		}

		if lastID > 0 {
			query.Where("operation.id < ?", lastID)
		}

		query.Limit(2) // It's a hack to avoid postgres "optimization". Limit = 1 is extremely slow.

		var ops []operation.Operation
		if err := storage.DB.Model().TableExpr("(?) as operation", query).
			ColumnExpr("operation.*").
			ColumnExpr("source.address as source__address").
			ColumnExpr("destination.address as destination__address").
			Join("LEFT JOIN accounts as source ON source.id = operation.source_id").
			Join("LEFT JOIN accounts as destination ON destination.id = operation.destination_id").
			Select(&ops); err != nil {
			return operation.Operation{}, err
		}
		if len(ops) > 0 {
			return ops[0], nil
		}

		current = current.AddDate(-1, 0, 0)
	}

	return operation.Operation{}, pg.ErrNoRows
}

// Get -
func (storage *Storage) Get(filters map[string]interface{}, size int64, sort bool) (operations []operation.Operation, err error) {
	query := storage.DB.Model((*operation.Operation)(nil)).Relation("Destination.address")

	for key, value := range filters {
		query.Where("? = ?", pg.Ident(key), value)
	}

	if sort {
		addOperationSorting(query)
	}

	if size > 0 {
		query.Limit(storage.GetPageSize(size))
	}

	err = query.Select(&operations)
	return operations, err
}

// GetByHash -
func (storage *Storage) GetByHash(hash []byte) (operations []operation.Operation, err error) {
	query := storage.DB.Model((*operation.Operation)(nil)).Where("hash = ?", hash)
	addOperationSorting(query)
	err = storage.DB.Model().TableExpr("(?) as operation", query).
		ColumnExpr("operation.*").
		ColumnExpr("source.address as source__address, source.alias as source__alias, source.type as source__type,source.id as source__id").
		ColumnExpr("destination.address as destination__address, destination.alias as destination__alias, destination.type as destination__type, destination.id as destination__id").
		Join("LEFT JOIN accounts as source ON source.id = operation.source_id").
		Join("LEFT JOIN accounts as destination ON destination.id = operation.destination_id").
		Select(&operations)
	return operations, err
}

// GetContract24HoursVolume -
func (storage *Storage) GetContract24HoursVolume(address string, entrypoints []string) (float64, error) {
	aDayAgo := time.Now().UTC().AddDate(0, 0, -1)
	var destinationID int64
	if err := storage.DB.Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Select(&destinationID); err != nil {
		return 0, err
	}

	var volume float64
	query := storage.DB.Model((*operation.Operation)(nil)).
		ColumnExpr("COALESCE(SUM(amount), 0)").
		Where("destination_id = ?", destinationID).
		Where("status = ?", types.OperationStatusApplied).
		Where("timestamp > ?", aDayAgo)

	if len(entrypoints) > 0 {
		query.WhereIn("entrypoint IN (?)", entrypoints)
	}

	err := query.Select(&volume)
	return volume, err
}

// GetByIDs -
func (storage *Storage) GetByIDs(ids ...int64) (result []operation.Operation, err error) {
	err = storage.DB.Model((*operation.Operation)(nil)).Where("id IN (?)", pg.In(ids)).Order("id asc").Select(&result)
	return
}

// GetByID -
func (storage *Storage) GetByID(id int64) (result operation.Operation, err error) {
	err = storage.DB.Model(&result).Relation("Destination").Where("operation.id = ?", id).First()
	return
}

// OPG -
func (storage *Storage) OPG(address string, size, lastID int64) ([]operation.OPG, error) {
	var accountID int64
	if err := storage.DB.Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Select(&accountID); err != nil {
		return nil, err
	}

	var (
		end        bool
		result     = make([]operation.OPG, 0)
		lastAction = time.Now().UTC()
		limit      = storage.GetPageSize(size)
	)

	lastActionSet := false
	if lastID > 0 {
		op, err := storage.GetByID(lastID)
		if err != nil {
			if !storage.IsRecordNotFound(err) {
				return nil, err
			}
		} else {
			lastAction = op.Timestamp
			lastActionSet = true
		}
	}
	if !lastActionSet && bcd.IsContractLazy(address) {
		if err := storage.DB.Model((*contract.Contract)(nil)).
			Column("last_action").
			Where("account_id = ?", accountID).
			Select(&lastAction); err != nil {
			return nil, err
		}
	}

	for !end {
		startTime, endTime, err := helpers.QuarterBoundaries(lastAction)
		if err != nil {
			return nil, err
		}

		subQuery := storage.DB.Model(new(operation.Operation)).
			Column("id", "hash", "counter", "status", "kind", "level", "timestamp", "content_index", "entrypoint").
			WhereGroup(
				func(q *orm.Query) (*orm.Query, error) {
					return q.Where("destination_id = ?", accountID).WhereOr("source_id = ?", accountID), nil
				},
			).
			Where("timestamp < ?", endTime).
			Where("timestamp >= ?", startTime).
			Order("id desc").
			Limit(1000)

		if lastID > 0 {
			subQuery.Where("id < ?", lastID)
		}

		var opg []operation.OPG
		if _, err := storage.DB.Query(&opg, `
		with opg as (?0)
		select 
			ta.last_id, 
			ta.status,
			ta.counter,
			ta.kind,
			ta.hash, 
			ta.level, 
			ta.timestamp, 
			ta.entrypoint, 
			ta.content_index,
			(select sum(case when source_id = ?1 then -"amount" else "amount" end) as "flow"
				from operations
				where hash = ta.hash and counter = ta.counter and (timestamp < ?4) and (timestamp >= ?3)
			) as "flow",
			(select sum(internal::integer) as internals
				from operations
				where hash = ta.hash and counter = ta.counter and (timestamp < ?4) and (timestamp >= ?3)
			),
			(select sum("burned") + sum("fee") as total_cost
				from operations
				where hash = ta.hash and counter = ta.counter and (timestamp < ?4) and (timestamp >= ?3)
			)
		from (
			select 
				min(id) as last_id, 
				hash, 
				counter, 
				max(status) as status, 
				min(kind) as kind, 
				min(level) as level, 
				min(timestamp) as timestamp, 
				min(content_index) as content_index,
				string_agg(entrypoint, ',') as entrypoint
			from opg
			group by hash, counter
			order by last_id desc
			limit ?2
		) as ta
		order by last_id desc
	`, subQuery, accountID, limit, startTime, endTime); err != nil {
			return nil, err
		}

		count := int(size) - len(result)
		if count < len(opg) {
			opg = opg[:count]
		}

		result = append(result, opg...)

		if len(result) < limit {
			lastAction = lastAction.AddDate(0, -3, 0)
			if lastAction.Before(consts.BeginningOfTime) {
				break
			}
		}

		end = len(result) == limit
	}

	return result, nil
}

// GetByHashAndCounter -
func (storage *Storage) GetByHashAndCounter(hash []byte, counter int64) ([]operation.Operation, error) {
	var operations []operation.Operation
	err := storage.DB.Model((*operation.Operation)(nil)).
		Where("hash = ?", hash).
		Where("counter = ?", counter).
		Relation("Destination").Relation("Source").Relation("Initiator").Relation("Delegate").
		Order("id asc").
		Select(&operations)
	return operations, err
}

// GetImplicitOperation -
func (storage *Storage) GetImplicitOperation(counter int64) (operation.Operation, error) {
	var op operation.Operation
	err := storage.DB.Model((*operation.Operation)(nil)).
		Where("hash is null").
		Where("counter = ?", counter).
		Relation("Destination").Relation("Source").Relation("Initiator").Relation("Delegate").
		Order("id asc").
		Select(&op)
	return op, err
}

// ListEvents -
func (storage *Storage) ListEvents(accountID int64, size, offset int64) ([]operation.Operation, error) {
	query := storage.DB.Model((*operation.Operation)(nil)).
		Where("source_id = ?", accountID).
		Where("kind = 7").
		Order("id desc")

	if offset > 0 {
		query.Offset(int(offset))
	}
	if size > 0 {
		query.Limit(int(size))
	} else {
		query.Limit(10)
	}

	var op []operation.Operation
	err := query.Select(&op)
	return op, err
}

// EventsCount -
func (storage *Storage) EventsCount(accountID int64) (int, error) {
	return storage.DB.Model((*operation.Operation)(nil)).
		Where("source_id = ?", accountID).
		Where("kind = 7").Count()
}

const cteContractStatsTemplate = `with last_operations as (
	SELECT timestamp
	FROM operations AS "operation" 
	WHERE ((destination_id = ?) OR (source_id = ?))
	order by timestamp desc
    FETCH NEXT 20 ROWS ONLY
)
select max(timestamp) as timestamp from last_operations`

type lastTimestampResult struct {
	Timestamp time.Time `pg:"timestamp"`
}

// ContractStats -
func (storage *Storage) ContractStats(address string) (stats operation.ContractStats, err error) {
	var accountID int64
	if err = storage.DB.Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Select(&accountID); err != nil {
		return
	}

	if bcd.IsContractLazy(address) {
		if err := storage.DB.Model((*contract.Contract)(nil)).
			Column("last_action").
			Where("account_id = ?", accountID).
			Select(&stats.LastAction); err != nil {
			return stats, err
		}
	} else {
		var result lastTimestampResult
		if _, err := storage.DB.QueryOne(&result, cteContractStatsTemplate, accountID, accountID); err != nil {
			return stats, err
		}
		stats.LastAction = result.Timestamp
	}

	count, err := storage.DB.Model((*operation.Operation)(nil)).WhereGroup(
		func(q *orm.Query) (*orm.Query, error) {
			return q.Where("destination_id = ?", accountID).WhereOr("source_id = ?", accountID), nil
		},
	).CountEstimate(100000)
	if err != nil {
		return
	}

	stats.Count = int64(count)

	return
}

// Origination -
func (storage *Storage) Origination(accountID int64) (operation.Operation, error) {
	var result operation.Operation
	err := storage.DB.Model((*operation.Operation)(nil)).
		Where("destination_id = ?", accountID).
		Where("kind = ?", types.OperationKindOrigination).
		Limit(1).
		Select(&result)
	return result, err
}

func addOperationSorting(query *orm.Query) {
	query.OrderExpr("operation.level desc, operation.counter desc, operation.id asc")
}
