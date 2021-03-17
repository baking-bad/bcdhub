package contract

import (
	"math/rand"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
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
func (storage *Storage) Get(network, address string) (response contract.Contract, err error) {
	err = storage.DB.Table(models.DocContracts).
		Scopes(core.NetworkAndAddress(network, address)).
		First(&response).Error
	return
}

// GetMany -
func (storage *Storage) GetMany(by map[string]interface{}) (response []contract.Contract, err error) {
	query := storage.DB.Table(models.DocContracts)
	query.Where(by)
	err = query.Find(&response).Error
	return
}

// GetRandom -
func (storage *Storage) GetRandom(network string) (response contract.Contract, err error) {
	queryCount := storage.DB.Table(models.DocContracts).Where("tx_count > 2")
	if network != "" {
		queryCount.Where("network = ?", network)
	}
	var count int64
	if err = queryCount.Count(&count).Error; err != nil {
		return
	}

	query := storage.DB.Table(models.DocContracts).Where("tx_count > 2")
	if network != "" {
		query.Where("network = ?", network)
	}
	err = query.Limit(1).Offset(rand.Intn(int(count))).First(&response).Error
	return
}

// IsFA -
func (storage *Storage) IsFA(network, address string) (bool, error) {
	var count int64
	err := storage.DB.Table(models.DocContracts).
		Scopes(core.NetworkAndAddress(network, address)).
		Where("tags IN ?", []string{"fa12", "fa1"}).
		Count(&count).
		Error

	return count > 0, err
}

// UpdateMigrationsCount -
func (storage *Storage) UpdateMigrationsCount(address, network string) error {
	return storage.DB.Raw(`UPDATE contracts SET migrations_count = migrations_count + 1 WHERE address = ? AND network = ?;`, address, network).Error
}

// GetAddressesByNetworkAndLevel -
func (storage *Storage) GetAddressesByNetworkAndLevel(network string, maxLevel int64) ([]string, error) {
	var addresses []string
	err := storage.DB.Table(models.DocContracts).
		Select("address").
		Where("network = ?", network).
		Where("level > ?", maxLevel).
		Find(&addresses).
		Error

	return addresses, err
}

// GetIDsByAddresses -
func (storage *Storage) GetIDsByAddresses(addresses []string, network string) ([]string, error) {
	var ids []string
	err := storage.DB.Table(models.DocContracts).
		Where("network = ?", network).
		Where("address IN ?", addresses).
		Pluck("id", &ids).
		Error

	return ids, err
}

// GetByAddresses -
func (storage *Storage) GetByAddresses(addresses []contract.Address) (response []contract.Contract, err error) {
	query := storage.DB.Table(models.DocContracts)

	if len(addresses) > 0 {
		subQuery := storage.DB.Where(
			storage.DB.Scopes(core.NetworkAndAddress(addresses[0].Network, addresses[0].Address)),
		)
		for i := 1; i < len(addresses); i++ {
			subQuery.Or(
				storage.DB.Scopes(core.NetworkAndAddress(addresses[i].Network, addresses[i].Address)),
			)
		}
		query.Where(subQuery)
	}

	err = query.Find(&response).Error
	return
}

// GetProjectsLastContract -
func (storage *Storage) GetProjectsLastContract(c *contract.Contract) (response []contract.Contract, err error) {
	subQuery := storage.DB.Table(models.DocContracts).Where(
		storage.DB.Where("fingerprint_code = ?", c.FingerprintCode).
			Where("fingerprint_parameter = ?", c.FingerprintParameter).
			Where("fingerprint_storage = ?", c.FingerprintStorage),
	)
	if c != nil {
		if c.Manager != "" {
			subQuery.Or("manager = ?", c.Manager)
		}
		if c.Language != "" {
			subQuery.Or("language = ?", c.Language)
		}
	}

	query := storage.DB.Table(models.DocContracts).
		Select("MAX(id) as id").
		Where("project_id != ''").
		Where(subQuery).
		Group("project_id")

	err = storage.DB.Table(models.DocContracts).Where("id IN (?)", query).Find(&response).Error
	return
}

// GetSameContracts -
func (storage *Storage) GetSameContracts(c contract.Contract, manager string, size, offset int64) (pcr contract.SameResponse, err error) {
	if c.FingerprintCode == "" || c.FingerprintParameter == "" || c.FingerprintStorage == "" {
		return pcr, errors.Wrap(consts.ErrInvalidFingerprint, c.Address)
	}

	limit := core.GetPageSize(size)

	query := storage.DB.Table(models.DocContracts).Where("hash = ?", c.Hash).Where("address != ?", c.Address)
	if manager != "" {
		query.Where("manager = ?", manager)
	}
	query.Order("last_action desc").Limit(limit).Offset(int(offset))
	if err = query.Find(&pcr.Contracts).Error; err != nil {
		return
	}

	countQuery := storage.DB.Table(models.DocContracts).Where("hash = ?", c.Hash).Where("address != ?", c.Address)
	if manager != "" {
		countQuery.Where("manager = ?", manager)
	}
	err = countQuery.Order("last_action desc").Count(&pcr.Count).Error
	return
}

// GetSimilarContracts -
func (storage *Storage) GetSimilarContracts(c contract.Contract, size, offset int64) ([]contract.Similar, int, error) {
	if c.FingerprintCode == "" || c.FingerprintParameter == "" || c.FingerprintStorage == "" {
		return nil, 0, errors.Wrap(consts.ErrInvalidFingerprint, c.Address)
	}

	limit := core.GetPageSize(size)

	subQuery := storage.DB.Table(models.DocContracts).
		Select("MAX(id) as id").
		Where("project_id = ?", c.ProjectID).
		Where("hash != ?", c.Hash).
		Group("hash")

	var pcr []contract.Similar
	if err := storage.DB.Table(models.DocContracts).
		Where("id IN (?)", subQuery).
		Order("last_action desc").
		Limit(limit).
		Offset(int(offset)).
		Find(&pcr).Error; err != nil {
		return nil, 0, err
	}

	var count int64
	err := storage.DB.Table(models.DocContracts).
		Where("project_id = ?", c.ProjectID).
		Where("hash != ?", c.Hash).
		Group("hash").
		Count(&count).
		Error

	return pcr, int(count), err
}

// GetDiffTasks -
func (storage *Storage) GetDiffTasks() ([]contract.DiffTask, error) {
	var contracts []contract.Contract
	query := storage.DB.Table(models.DocContracts).Group("project_id, hash").Order("last_action DESC")
	if err := query.Find(&contracts).Error; err != nil {
		return nil, err
	}

	tasks := make([]contract.DiffTask, 0)
	for i, first := range contracts {
		for j := i; j < len(contracts); j++ {
			second := contracts[j]
			if first.ProjectID != second.ProjectID {
				continue
			}

			tasks = append(tasks, contract.DiffTask{
				Network1: first.Network,
				Address1: first.Address,
				Network2: second.Network,
				Address2: second.Address,
			})
		}
	}

	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(tasks), func(i, j int) { tasks[i], tasks[j] = tasks[j], tasks[i] })
	return tasks, nil
}

// GetTokens -
func (storage *Storage) GetTokens(network, tokenInterface string, offset, size int64) ([]contract.Contract, int64, error) {
	tags := []string{"fa12", "fa1", "fa2"}
	if tokenInterface == "fa12" || tokenInterface == "fa1" || tokenInterface == "fa2" {
		tags = []string{tokenInterface}
	}

	var contracts []contract.Contract
	err := storage.DB.Table(models.DocContracts).
		Scopes(core.Network(network)).
		Where("tags IN ?", tags).
		Order("timestamp desc").
		Find(&contracts).
		Error
	if err != nil {
		return nil, 0, err
	}

	var count int64
	err = storage.DB.Table(models.DocContracts).
		Scopes(core.Network(network)).
		Where("tags IN ?", tags).
		Count(&count).
		Error

	return contracts, count, err
}

// UpdateField -
func (storage *Storage) UpdateField(where []contract.Contract, fields ...string) error {
	if len(where) == 0 {
		return nil
	}

	// For a deadlock reason don't wrap the requests to transaction
	for i := range where {
		updates := core.GetFieldsForModel(where[i], fields...)

		if err := storage.DB.Table(models.DocContracts).
			Where("id = ?", where[i].ID).
			Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil

}

// Stats -
func (storage *Storage) Stats(c contract.Contract) (stats contract.Stats, err error) {
	err = storage.DB.Table(models.DocContracts).
		Where("hash = ?", c.Hash).
		Where("address != ?", c.Address).
		Count(&stats.SameCount).
		Error
	if err != nil {
		return
	}

	err = storage.DB.Table(models.DocContracts).
		Where("project_id = ?", c.ProjectID).
		Where("hash != ?", c.Hash).
		Group("hash").
		Count(&stats.SimilarCount).
		Error
	return
}

// GetByIDs -
func (storage *Storage) GetByIDs(ids ...int64) (result []contract.Contract, err error) {
	err = storage.DB.Table(models.DocContracts).Order("id asc").Find(&result, ids).Error
	return
}
