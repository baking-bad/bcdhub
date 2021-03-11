package operation

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/pkg/errors"
	"github.com/restream/reindexer"
)

const (
	sortString = "level * 10000000000 + counter * 1000 + internal ? (998 - nonce) : 999"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

func (storage *Storage) getContractOPG(address, network string, size uint64, filters map[string]interface{}) ([]opgForContract, error) {
	if size == 0 {
		size = core.DefaultSize
	}

	query := storage.db.Query(models.DocOperations).
		Distinct("hash").Distinct("counter").Distinct("level").
		OpenBracket().
		Match("source", address).
		Or().
		Match("destination", address).
		CloseBracket().
		Match("network", network)

	if err := prepareOperationFilters(filters, query); err != nil {
		return nil, err
	}
	it := query.Limit(int(size)).Sort("level", true).Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	resp := make([]opgForContract, 0)
	for it.Next() {
		var obj opgForContract
		it.NextObj(&obj)
		resp = append(resp, obj)
	}

	return resp, nil
}

func prepareOperationFilters(filters map[string]interface{}, query *reindexer.Query) error {
	for k, v := range filters {
		if v == "" || v == nil {
			continue
		}
		switch k {
		case "from":
			query = query.Where("timestamp", reindexer.GE, v)
		case "to":
			query = query.Where("timestamp", reindexer.LE, v)
		case "entrypoints":
			query = query.Where("entrypoint", reindexer.EQ, v)
		case "last_id":
			query = query.Where("timestamp", reindexer.LT, v)
		case "status":
			query = query.Where("status", reindexer.EQ, v)
		default:
			return errors.Errorf("Unknown operation filter: %s %v", k, v)
		}

	}
	return nil
}

// GetByContract -
func (storage *Storage) GetByContract(network, address string, size uint64, filters map[string]interface{}) (po operation.Pageable, err error) {
	opg, err := storage.getContractOPG(address, network, size, filters)
	if err != nil {
		return
	}
	if len(opg) == 0 {
		return
	}

	query := storage.db.Query(models.DocOperations).
		Match("network", network).
		OpenBracket()

	for i := range opg {
		query = query.OpenBracket().Match("hash", opg[i].Hash).WhereInt64("counter", reindexer.EQ, opg[i].Counter).CloseBracket()
		if len(opg)-1 > i {
			query = query.Or()
		}
	}
	query = query.CloseBracket().Sort(sortString, true)
	query.AggregateMin("indexed_time")

	it := query.Exec()
	defer it.Close()

	if err = it.Error(); err != nil {
		return
	}

	po.Operations = make([]operation.Operation, it.Count())
	for i := 0; i < it.Count(); i++ {
		it.NextObj(&po.Operations[i])
	}
	po.LastID = fmt.Sprintf("%.0f", it.AggResults()[0].Value)
	return
}

// Last -
func (storage *Storage) Last(network, address string, indexedTime int64) (op operation.Operation, err error) {
	query := storage.db.Query(models.DocOperations).
		Match("destination", address).
		Match("network", network).
		Match("status", consts.Applied).
		WhereString("deffated_storage", reindexer.EMPTY, "").
		WhereInt64("indexed_time", reindexer.LT, indexedTime).
		Sort("indexed_time", true)

	err = storage.db.GetOne(query, &op)
	return
}

// Get -
func (storage *Storage) Get(filters map[string]interface{}, size int64, sort bool) (operations []operation.Operation, err error) {
	query := storage.db.Query(models.DocOperations)
	for field, value := range filters {
		query = query.Where(field, reindexer.EQ, value)
	}

	if sort {
		query = query.Sort(sortString, true)
	}

	if size > 0 {
		query = query.Limit(int(size))
	}

	err = storage.db.GetAllByQuery(query, &operations)
	return
}

// GetStats -
func (storage *Storage) GetStats(network, address string) (stats operation.Stats, err error) {
	query := storage.db.Query(models.DocOperations).
		Distinct("hash").
		Match("network", network).
		OpenBracket().
		Match("source", address).
		Or().
		Match("destination", address).
		CloseBracket().
		ReqTotal()

	query.AggregateMax("timestamp")

	it := query.Exec()
	defer it.Close()

	if err = it.Error(); err != nil {
		return
	}

	stats.Count = int64(it.TotalCount())
	stats.LastAction = time.Unix(int64(it.AggResults()[0].Value), 0) // TODO: is the date valid? check parsing`
	return
}

// GetTokensStats -
func (storage *Storage) GetTokensStats(network string, addresses, entrypoints []string) (map[string]operation.TokenUsageStats, error) {
	query := storage.db.Query(models.DocOperations)

	if len(addresses) > 0 {
		query = query.Match("destination", addresses...)
	}
	if len(entrypoints) > 0 {
		query = query.Match("entrypoint", entrypoints...)
	}

	operations := make([]operation.Operation, 0)
	if err := storage.db.GetAllByQuery(query, &operations); err != nil {
		return nil, err
	}

	all := make(map[string][]int64)
	for i := range operations {
		id := fmt.Sprintf("%s|%s", operations[i].Destination, operations[i].Entrypoint)
		if _, ok := all[id]; !ok {
			all[id] = make([]int64, 0)
		}
		all[id] = append(all[id], operations[i].Result.ConsumedGas)
	}

	usageStats := make(map[string]operation.TokenUsageStats)
	for id, arr := range all {
		parts := strings.Split(id, "|")
		var total float64
		for _, value := range arr {
			total += float64(value)
		}
		avg := int64(total / float64(len(arr)))
		address := parts[0]
		entrypoint := parts[1]

		usage := operation.TokenMethodUsageStats{
			ConsumedGas: avg,
			Count:       int64(len(arr)),
		}
		if _, ok := usageStats[address]; !ok {
			usageStats[address] = make(operation.TokenUsageStats)
		}
		usageStats[address][entrypoint] = usage
	}
	return usageStats, nil
}

// GetParticipatingContracts -
func (storage *Storage) GetParticipatingContracts(network string, fromLevel, toLevel int64) ([]string, error) {
	it := storage.db.Query(models.DocOperations).
		Select("destination", "source").
		Match("network", network).
		WhereInt64("level", reindexer.LE, fromLevel).
		WhereInt64("level", reindexer.GT, toLevel).Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	exists := make(map[string]struct{})
	addresses := make([]string, 0)

	type response struct {
		Source      string `reindex:"source"`
		Destination string `reindex:"destination"`
	}
	for it.Next() {
		var item response
		it.NextObj(&item)
		if _, ok := exists[item.Destination]; bcd.IsContract(item.Destination) && !ok {
			exists[item.Destination] = struct{}{}
			addresses = append(addresses, item.Destination)
		}
		if _, ok := exists[item.Source]; bcd.IsContract(item.Source) && !ok {
			exists[item.Source] = struct{}{}
			addresses = append(addresses, item.Source)
		}
	}

	return addresses, nil
}

// RecalcStats -
func (storage *Storage) RecalcStats(network, address string) (stats operation.ContractStats, err error) {
	query := storage.db.Query(models.DocOperations).
		Match("network", network).
		Match("status", consts.Applied).
		OpenBracket().
		Match("source", address).
		Or().
		Match("destination", address).
		CloseBracket().
		ReqTotal()

	query.AggregateMax("timestamp")

	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return stats, it.Error()
	}

	stats.TxCount = int64(it.TotalCount())
	stats.LastAction = time.Unix(0, int64(it.AggResults()[0].Value)*1000000).UTC()

	for it.Next() {
		var op operation.Operation
		it.NextObj(&op)
		if op.Source == address {
			stats.Balance -= op.Amount
		} else {
			stats.Balance += op.Amount
		}
	}
	return
}

// GetDAppStats -
func (storage *Storage) GetDAppStats(network string, addresses []string, period string) (stats operation.DAppStats, err error) {
	query := storage.db.Query(models.DocOperations).
		Distinct("source").
		Match("network", network).
		Match("status", consts.Applied).
		Match("destination", addresses...).
		Not().
		Match("entrypoint", "")

	if err = periodToRange(period, query); err != nil {
		return
	}
	query = query.ReqTotal()
	query.AggregateSum("amount")

	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return stats, it.Error()
	}

	stats.Calls = int64(it.TotalCount())
	stats.Users = int64(len(it.AggResults()[0].Distincts))
	stats.Volume = int64(it.AggResults()[1].Value)
	return
}

func periodToRange(period string, query *reindexer.Query) error {
	now := time.Now()
	switch period {
	case "year":
		query.WhereInt64("timestamp", reindexer.GT, now.AddDate(-1, 0, 0).Unix())
	case "month":
		query.WhereInt64("timestamp", reindexer.GT, now.AddDate(0, -1, 0).Unix())
	case "week":
		query.WhereInt64("timestamp", reindexer.GT, now.AddDate(0, 0, -7).Unix())
	case "day":
		query.WhereInt64("timestamp", reindexer.GT, now.AddDate(0, 0, -1).Unix())
	case "all":
		return nil
	default:
		return errors.Errorf("Unknown period value: %s", period)
	}
	return nil
}

// GetContract24HoursVolume -
func (storage *Storage) GetContract24HoursVolume(network, address string, entrypoints []string) (float64, error) {
	return 0, nil
}
