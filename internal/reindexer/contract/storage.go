package contract

import (
	"math/rand"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/pkg/errors"
	"github.com/restream/reindexer"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// Get -
func (storage *Storage) Get(by map[string]interface{}) (c contract.Contract, err error) {
	query := storage.db.Query(models.DocContracts)

	for field, value := range by {
		query = query.Where(field, reindexer.EQ, value)
	}

	err = storage.db.GetOne(query, &c)
	return
}

// GetMany -
func (storage *Storage) GetMany(by map[string]interface{}) (contracts []contract.Contract, err error) {
	query := storage.db.Query(models.DocContracts)

	for field, value := range by {
		query = query.Where(field, reindexer.EQ, value)
	}

	err = storage.db.GetAllByQuery(query, &contracts)
	return
}

// GetRandom -
func (storage *Storage) GetRandom() (c contract.Contract, err error) {
	query := storage.db.Query(models.DocContracts).WhereInt("tx_count", reindexer.GE, 2)
	count, err := storage.db.Count(query)
	if err != nil {
		return c, err
	}

	idx := rand.Intn(int(count))
	secondQuery := storage.db.Query(models.DocContracts).WhereInt("tx_count", reindexer.GE, 2).Limit(1).Offset(idx)
	err = storage.db.GetOne(secondQuery, &c)
	return
}

// IsFA -
func (storage *Storage) IsFA(network, address string) (bool, error) {
	query := storage.db.Query(models.DocContracts).
		Match("network", network).
		Match("address", address).
		Match("tags", "fa12", "fa1")

	it := query.Exec()
	if it.Error() != nil {
		return false, it.Error()
	}

	return it.TotalCount() == 1, nil
}

// UpdateMigrationsCount -
func (storage *Storage) UpdateMigrationsCount(address, network string) error {
	contract := contract.NewEmptyContract(network, address)
	it := storage.db.Query(models.DocOperations).
		Where("id", reindexer.EQ, contract.GetID()).
		Set("migrations_count", "migrations_count + 1").
		Update()
	defer it.Close()

	return it.Error()
}

// GetAddressesByNetworkAndLevel -
func (storage *Storage) GetAddressesByNetworkAndLevel(network string, maxLevel int64) ([]string, error) {
	query := storage.db.Query(models.DocContracts).
		Select("address").
		Match("network", network).
		WhereInt64("level", reindexer.GT, maxLevel)

	addresses := make([]string, 0)
	err := storage.db.GetAllByQuery(query, &addresses)
	return addresses, err
}

// GetIDsByAddresses -
func (storage *Storage) GetIDsByAddresses(addresses []string, network string) ([]string, error) {
	if len(addresses) == 0 {
		return nil, nil
	}

	query := storage.db.Query(models.DocContracts).
		Select("id").
		Match("network", network).
		OpenBracket()

	for i := range addresses {
		query = query.Match("address", addresses[i])
		if i < len(addresses)-1 {
			query = query.Or()
		}
	}
	query = query.CloseBracket()

	ids := make([]string, 0)
	err := storage.db.GetAllByQuery(query, &ids)
	return ids, err
}

// GetByAddresses -
func (storage *Storage) GetByAddresses(addresses []contract.Address) (contracts []contract.Contract, err error) {
	if len(addresses) == 0 {
		return
	}
	query := storage.db.Query(models.DocContracts)
	for i := range addresses {
		query = query.OpenBracket().
			Match("address", addresses[i].Address).
			Match("network", addresses[i].Network).
			CloseBracket()
		if i < len(addresses)-1 {
			query = query.Or()
		}
	}

	err = storage.db.GetAllByQuery(query, &contracts)
	return
}

// GetProjectsLastContract -
func (storage *Storage) GetProjectsLastContract() ([]contract.Contract, error) {
	query := storage.db.Query(models.DocContracts).Sort("timestamp", true)
	return storage.topContracts(query, func(c contract.Contract) string {
		return c.ProjectID
	})
}

// GetSameContracts -
func (storage *Storage) GetSameContracts(c contract.Contract, size, offset int64) (pcr contract.SameResponse, err error) {
	if c.Fingerprint == nil {
		return pcr, errors.Errorf("Invalid contract data")
	}

	if size == 0 {
		size = core.DefaultSize
	}

	query := storage.db.Query(models.DocContracts).
		Match("hash", c.Hash).
		Match("address", c.Address).
		Limit(int(size)).
		Offset(int(offset)).
		Sort("last_action", true)

	contracts := make([]contract.Contract, 0)
	total, err := storage.db.GetAllByQueryWithTotal(query, &contracts)
	if err != nil {
		return
	}

	pcr.Contracts = contracts
	pcr.Count = int64(total)
	return
}

// GetSimilarContracts -
func (storage *Storage) GetSimilarContracts(c contract.Contract, size, offset int64) ([]contract.Similar, int, error) {
	if c.Fingerprint == nil {
		return nil, 0, nil
	}

	if size == 0 {
		size = core.DefaultSize
	}

	query := storage.db.Query(models.DocContracts).
		Distinct("hash").
		Select("hash").
		Match("project_id", c.ProjectID).
		Not().
		Match("hash", c.Hash).ReqTotal()

	query.AggregateFacet("hash").Limit(int(size)).Offset(int(offset))

	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, 0, it.Error()
	}

	count := make(map[string]int)
	hash := make([]string, 0)
	agg := it.AggResults()[1]
	for _, bucket := range agg.Facets {
		hash = append(hash, bucket.Values[0])
		count[bucket.Values[0]] = bucket.Count
	}

	sit := storage.db.Query(models.DocContracts).Match("hash", hash...).Exec()
	defer it.Close()

	if sit.Error() != nil {
		return nil, 0, sit.Error()
	}

	found := make(map[string]struct{})
	contracts := make([]contract.Similar, 0)

	for sit.Next() {
		var c contract.Contract
		sit.NextObj(&c)
		if _, ok := found[c.Hash]; ok {
			continue
		}
		found[c.Hash] = struct{}{}
		contracts = append(contracts, contract.Similar{
			Contract: &c,
			Count:    int64(count[c.Hash]),
		})

	}

	total := len(it.AggResults()[0].Distincts)
	return contracts, total, nil
}

// GetDiffTasks -
func (storage *Storage) GetDiffTasks() ([]contract.DiffTask, error) {
	return nil, nil
}

// GetTokens -
func (storage *Storage) GetTokens(network, tokenInterface string, offset, size int64) ([]contract.Contract, int64, error) {
	tags := []string{"fa12", "fa1", "fa2"}
	if tokenInterface == "fa12" || tokenInterface == "fa1" || tokenInterface == "fa2" {
		tags = []string{tokenInterface}
	}

	query := storage.db.Query(models.DocContracts).
		Match("network", network).
		Match("tags", tags...).
		Sort("timestamp", true)

	if size > 0 {
		query = query.Limit(int(size))
	}

	if offset > 0 {
		query = query.Offset(int(offset))
	}

	contracts := make([]contract.Contract, 0)
	total, err := storage.db.GetAllByQueryWithTotal(query, &contracts)
	if err != nil {
		return nil, 0, err
	}

	return contracts, int64(total), nil
}

func (storage *Storage) topContracts(query *reindexer.Query, idFunc func(c contract.Contract) string) ([]contract.Contract, error) {
	all := make([]contract.Contract, 0)
	if err := storage.db.GetAllByQuery(query, &all); err != nil {
		return nil, err
	}

	response := make([]contract.Contract, 0)
	found := make(map[string]struct{})
	for i := range all {
		id := idFunc(all[i])
		if _, ok := found[id]; ok {
			continue
		}
		found[id] = struct{}{}
		response = append(response, all[i])
	}
	return response, nil

}

// UpdateField -
func (storage *Storage) UpdateField(where []contract.Contract, fields ...string) error {
	if len(where) == 0 {
		return nil
	}
	tx, err := storage.db.BeginTx(models.DocContracts)
	if err != nil {
		return err
	}
	for i := range where {
		query := tx.Query().Match("id", where[i].GetID())
		for j := range fields {
			value := storage.db.GetFieldValue(where[i], fields[j])
			query = query.Set(fields[j], value)
		}
		it := query.Update()
		defer it.Close()

		if it.Error() != nil {
			return it.Error()
		}
	}
	return tx.Commit()
}
