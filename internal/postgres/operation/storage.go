package operation

import (
	"context"
	"database/sql"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/uptrace/bun"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(es *core.Postgres) *Storage {
	return &Storage{es}
}

// Last - get last operation by `filters` with not empty deffated_storage
func (storage *Storage) Last(ctx context.Context, filters map[string]interface{}, lastID int64) (operation.Operation, error) {
	var (
		current = time.Now()
		endTime = consts.BeginningOfTime
	)

	if val, ok := filters["timestamp"]; ok {
		if tf, ok := val.(core.TimestampFilter); ok {
			switch {
			case !tf.Lt.IsZero():
				current = tf.Lt
			case !tf.Lte.IsZero():
				current = tf.Lte
			}

			switch {
			case !tf.Gt.IsZero():
				endTime = tf.Gt
			case !tf.Gte.IsZero():
				endTime = tf.Gte
			}
		}
	}

	for current.After(endTime) {
		query := storage.DB.NewSelect().Model((*operation.Operation)(nil)).
			Where("deffated_storage is not null").
			OrderExpr("operation.id desc")

		for key, value := range filters {
			switch val := value.(type) {
			case core.TimestampFilter:
				query = val.Apply(query)
			default:
				query.Where("? = ?", bun.Ident(key), value)
			}
		}

		lowCurrent := current.AddDate(0, -3, 0)
		query.
			Where("timestamp >= ?", lowCurrent).
			Where("timestamp < ?", current)

		if lastID > 0 {
			query.Where("operation.id < ?", lastID)
		}

		query.Limit(1)

		var ops []operation.Operation
		if err := storage.DB.NewSelect().TableExpr("(?) as operation", query).
			ColumnExpr("operation.*").
			ColumnExpr("source.address as source__address").
			ColumnExpr("destination.address as destination__address").
			Join("LEFT JOIN accounts as source ON source.id = operation.source_id").
			Join("LEFT JOIN accounts as destination ON destination.id = operation.destination_id").
			Scan(ctx, &ops); err != nil {
			return operation.Operation{}, err
		}
		if len(ops) > 0 {
			return ops[0], nil
		}

		current = lowCurrent
	}

	return operation.Operation{}, sql.ErrNoRows
}

// GetByID -
func (storage *Storage) GetByID(ctx context.Context, id int64) (result operation.Operation, err error) {
	err = storage.DB.NewSelect().
		Model(&result).
		Relation("Destination").
		Where("operation.id = ?", id).
		Limit(1).
		Scan(ctx)
	return
}

// OPG -
func (storage *Storage) OPG(ctx context.Context, address string, size, lastID int64) ([]operation.OPG, error) {
	var accountID int64
	if err := storage.DB.NewSelect().
		Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Scan(ctx, &accountID); err != nil {
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
		op, err := storage.GetByID(ctx, lastID)
		if err != nil {
			if !storage.IsRecordNotFound(err) {
				return nil, err
			}
		} else {
			lastAction = op.Timestamp
			lastActionSet = true
		}
	}
	if !lastActionSet {
		if err := storage.DB.NewSelect().
			Model((*account.Account)(nil)).
			Column("last_action").
			Where("id = ?", accountID).
			Scan(ctx, &lastAction); err != nil {
			return nil, err
		}
	}

	for !end {
		startTime, endTime, err := helpers.QuarterBoundaries(lastAction)
		if err != nil {
			return nil, err
		}

		subQuery := storage.DB.NewSelect().Model(new(operation.Operation)).
			Column("id", "hash", "counter", "status", "kind", "level", "timestamp", "content_index", "entrypoint").
			WhereGroup(" AND ",
				func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("destination_id = ?", accountID).WhereOr("source_id = ?", accountID)
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
		if err := storage.DB.NewRaw(`
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
	`, subQuery, accountID, limit, startTime, endTime).Scan(ctx, &opg); err != nil {
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

// GetByHash -
func (storage *Storage) GetByHash(ctx context.Context, hash []byte) (operations []operation.Operation, err error) {
	query := storage.DB.NewSelect().Model((*operation.Operation)(nil)).
		Where("hash = ?", hash)

	addOperationSorting(query)
	err = storage.DB.NewSelect().TableExpr("(?) as operation", query).
		ColumnExpr("operation.*").
		ColumnExpr("source.address as source__address, source.type as source__type,source.id as source__id").
		ColumnExpr("destination.address as destination__address, destination.type as destination__type, destination.id as destination__id").
		Join("LEFT JOIN accounts as source ON source.id = operation.source_id").
		Join("LEFT JOIN accounts as destination ON destination.id = operation.destination_id").
		Scan(ctx, &operations)
	return operations, err
}

// GetByHashAndCounter -
func (storage *Storage) GetByHashAndCounter(ctx context.Context, hash []byte, counter int64) (operations []operation.Operation, err error) {
	query := storage.DB.NewSelect().Model(&operations)

	if hash == nil {
		query = query.Where("hash is null")
	} else {
		query = query.Where("hash = ?", hash)
	}

	err = query.Where("counter = ?", counter).
		Relation("Destination").Relation("Source").Relation("Initiator").Relation("Delegate").
		Order("id asc").
		Scan(ctx)
	return
}

// ListEvents -
func (storage *Storage) ListEvents(ctx context.Context, accountID int64, size, offset int64) (operations []operation.Operation, err error) {
	query := storage.DB.NewSelect().Model(&operations).
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

	err = query.Scan(ctx)
	return
}

// Origination -
func (storage *Storage) Origination(ctx context.Context, accountID int64) (result operation.Operation, err error) {
	err = storage.DB.NewSelect().
		Model(&result).
		Where("destination_id = ?", accountID).
		Where("kind = 2").
		Limit(1).
		Scan(ctx)
	return result, err
}

func addOperationSorting(query *bun.SelectQuery) {
	query.OrderExpr("operation.level desc, operation.counter desc, operation.id asc")
}
