package operation

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
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
	Hash    string
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

	query := storage.DB.Model((*operation.Operation)(nil)).Where("operation.network = ?", acc.Network).WhereGroup(func(q *orm.Query) (*orm.Query, error) {
		for i := range opg {
			q.WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				q.Where("operation.hash = ?", opg[i].Hash).Where("operation.counter = ?", opg[i].Counter)
				return q, nil
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
func (storage *Storage) Last(filters map[string]interface{}, lastID int64) (op operation.Operation, err error) {
	query := storage.DB.Model((*operation.Operation)(nil)).Relation("Source.address").Relation("Destination.address").Where("deffated_storage is not null")

	for key, value := range filters {
		query.Where("? = ?", pg.Ident(key), value)
	}

	if lastID > 0 {
		query.Where("operation.id < ?", lastID)
	}

	err = query.Order("operation.id desc").Limit(1).Select(&op)
	return
}

// Get -
func (storage *Storage) Get(filters map[string]interface{}, size int64, sort bool) (operations []operation.Operation, err error) {
	query := storage.DB.Model((*operation.Operation)(nil)).Relation("Source.address").Relation("Destination.address")

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

// GetContract24HoursVolume -
func (storage *Storage) GetContract24HoursVolume(network types.Network, address string, entrypoints []string) (float64, error) {
	aDayAgo := time.Now().UTC().AddDate(0, 0, -1)
	var destinationID int64
	if err := storage.DB.Model((*account.Account)(nil)).
		Column("id").
		Where("network = ?", network).
		Where("address = ?", address).
		Select(&destinationID); err != nil {
		return 0, err
	}

	var volume float64
	query := storage.DB.Model((*operation.Operation)(nil)).
		ColumnExpr("COALESCE(SUM(amount), 0)").
		Where("destination_id = ?", destinationID).
		Where("operation.network = ?", network).
		Where("status = ?", types.OperationStatusApplied).
		Where("timestamp > ?", aDayAgo)

	if len(entrypoints) > 0 {
		query.WhereIn("entrypoint IN (?)", entrypoints)
	}

	err := query.Select(&volume)
	return volume, err
}

type tokenStats struct {
	Address    string `pg:"destination__address"`
	Entrypoint string
	Gas        int64
	Count      int64
}

// GetTokensStats -
func (storage *Storage) GetTokensStats(network types.Network, addresses, entrypoints []string) (map[string]operation.TokenUsageStats, error) {
	var stats []tokenStats
	query := storage.DB.Model((*operation.Operation)(nil)).
		ColumnExpr("operation.entrypoint as entrypoint, COUNT(*) as count, SUM(consumed_gas) AS gas").
		Where("operation.network = ?", network).Relation("Destination.address")

	if len(addresses) > 0 {
		query.WhereIn("destination.address IN (?)", addresses)
	}

	if len(entrypoints) > 0 {
		query.WhereIn("operation.entrypoint IN (?)", entrypoints)
	}

	query.GroupExpr("destination.address, operation.entrypoint")

	if err := query.Select(&stats); err != nil {
		return nil, err
	}

	usageStats := make(map[string]operation.TokenUsageStats)
	for i := range stats {
		usage := operation.TokenMethodUsageStats{
			Count:       stats[i].Count,
			ConsumedGas: stats[i].Gas,
		}
		if _, ok := usageStats[stats[i].Address]; !ok {
			usageStats[stats[i].Address] = make(operation.TokenUsageStats)
		}
		usageStats[stats[i].Address][stats[i].Entrypoint] = usage
	}

	return usageStats, nil
}

// GetByIDs -
func (storage *Storage) GetByIDs(ids ...int64) (result []operation.Operation, err error) {
	err = storage.DB.Model().Table(models.DocOperations).Where("id IN (?)", pg.In(ids)).Order("id asc").Select(&result)
	return
}

// GetByID -
func (storage *Storage) GetByID(id int64) (result operation.Operation, err error) {
	err = storage.DB.Model(&result).Relation("Destination").Where("operation.id = ?", id).First()
	return
}

// GetDAppStats -
func (storage *Storage) GetDAppStats(network types.Network, addresses []string, period string) (stats operation.DAppStats, err error) {
	var ids []int64
	if len(addresses) > 0 {
		if err = storage.DB.Model((*account.Account)(nil)).
			Column("id").
			WhereIn("address IN (?)", addresses).
			Select(&ids); err != nil {
			return
		}
	}

	query, err := getDAppQuery(storage.DB, network, ids, period)
	if err != nil {
		return
	}

	if err = query.ColumnExpr("COUNT(*) as calls, SUM(amount) as volume").Select(&stats); err != nil {
		return
	}

	queryCount, err := getDAppQuery(storage.DB, network, ids, period)
	if err != nil {
		return
	}

	count, err := queryCount.Column("source_id").Group("source_id").Count()
	if err != nil {
		return
	}
	stats.Users = int64(count)
	return
}

func getDAppQuery(db pg.DBI, network types.Network, ids []int64, period string) (*orm.Query, error) {
	query := db.Model((*operation.Operation)(nil)).
		Where("network = ?", network).
		Where("status = ?", types.OperationStatusApplied)

	if len(ids) > 0 {
		query.WhereIn("destination_id IN (?)", ids)
	}

	err := periodToRange(query, period)
	return query, err
}

func periodToRange(query *orm.Query, period string) error {
	now := time.Now().UTC()
	switch period {
	case "year":
		now = now.AddDate(-1, 0, 0)
	case "month":
		now = now.AddDate(0, -1, 0)
	case "week":
		now = now.AddDate(0, 0, -7)
	case "day":
		now = now.AddDate(0, 0, -1)
	case "all":
		now = consts.BeginningOfTime
	default:
		return errors.Errorf("Unknown period value: %s", period)
	}
	query.Where("timestamp > ?", now)
	return nil
}

func addOperationSorting(query *orm.Query) {
	query.OrderExpr("operation.level desc, operation.counter desc, operation.id asc")
}
