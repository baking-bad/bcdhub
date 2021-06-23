package contract

import (
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
func (storage *Storage) Get(network types.Network, address string) (response contract.Contract, err error) {
	err = storage.DB.Scopes(core.NetworkAndAddress(network, address)).First(&response).Error
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
func (storage *Storage) GetRandom(network types.Network) (response contract.Contract, err error) {
	queryCount := storage.DB.Table(models.DocContracts).Where("tx_count > 2")
	if network != types.Empty {
		queryCount.Where("network = ?", network)
	}
	var count int64
	if err = queryCount.Count(&count).Error; err != nil {
		return
	}

	if count == 0 {
		return response, gorm.ErrRecordNotFound
	}

	query := storage.DB.Table(models.DocContracts).Where("tx_count > 2")
	if network != types.Empty {
		query.Where("network = ?", network)
	}
	err = query.Limit(1).Offset(rand.Intn(int(count))).First(&response).Error
	return
}

// UpdateMigrationsCount -
func (storage *Storage) UpdateMigrationsCount(network types.Network, address string) error {
	return storage.DB.Raw(`UPDATE contracts SET migrations_count = migrations_count + 1 WHERE address = ? AND network = ?;`, address, network).Error
}

// GetAddressesByNetworkAndLevel -
func (storage *Storage) GetAddressesByNetworkAndLevel(network types.Network, maxLevel int64) ([]string, error) {
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
func (storage *Storage) GetIDsByAddresses(network types.Network, addresses []string) ([]string, error) {
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
func (storage *Storage) GetProjectsLastContract(c contract.Contract, size, offset int64) (response []contract.Contract, err error) {
	if c.FingerprintCode == nil || c.FingerprintParameter == nil || c.FingerprintStorage == nil {
		return nil, nil
	}

	code := hex.EncodeToString(c.FingerprintCode)
	params := hex.EncodeToString(c.FingerprintParameter)
	s := hex.EncodeToString(c.FingerprintStorage)

	subQuery := storage.DB.Table(models.DocContracts).Where(
		storage.DB.Where("encode(fingerprint_code, 'hex') = ?", code).
			Where("encode(fingerprint_parameter, 'hex') = ?", params).
			Where("encode(fingerprint_storage, 'hex') = ?", s),
	)
	if c.Manager != "" {
		subQuery.Or("manager = ?", c.Manager)
	}
	if c.Language != "unknown" {
		subQuery.Or("language = ?", c.Language)
	}

	limit := storage.GetPageSize(size)

	query := storage.DB.Table(models.DocContracts).
		Select("MAX(id) as id").
		Where("project_id != ''").
		Where(subQuery).
		Group("project_id").
		Limit(limit).
		Offset(int(offset)).
		Order("id desc")

	err = storage.DB.Table(models.DocContracts).Where("id IN (?)", query).Find(&response).Error
	return
}

// GetSameContracts -
func (storage *Storage) GetSameContracts(c contract.Contract, manager string, size, offset int64) (pcr contract.SameResponse, err error) {
	if c.FingerprintCode == nil || c.FingerprintParameter == nil || c.FingerprintStorage == nil {
		return pcr, errors.Wrap(consts.ErrInvalidFingerprint, c.Address)
	}

	limit := storage.GetPageSize(size)

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
	if c.FingerprintCode == nil || c.FingerprintParameter == nil || c.FingerprintStorage == nil {
		return nil, 0, errors.Wrap(consts.ErrInvalidFingerprint, c.Address)
	}

	limit := storage.GetPageSize(size)

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
func (storage *Storage) GetTokens(network types.Network, tokenInterface string, offset, size int64) ([]contract.Contract, int64, error) {
	tags := types.FA12Tag | types.FA1Tag | types.FA2Tag
	if tokenInterface == "fa1-2" || tokenInterface == "fa1" || tokenInterface == "fa2" {
		tags = types.NewTags([]string{tokenInterface})
	}

	var contracts []contract.Contract
	err := storage.DB.Table(models.DocContracts).
		Scopes(core.Network(network)).
		Where("(tags & ?) > 0", tags).
		Order("id desc").
		Limit(storage.GetPageSize(size)).
		Offset(int(offset)).
		Find(&contracts).
		Error
	if err != nil {
		return nil, 0, err
	}

	var count int64
	err = storage.DB.Table(models.DocContracts).
		Scopes(core.Network(network)).
		Where("(tags & ?) > 0", tags).
		Count(&count).
		Error

	return contracts, count, err
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

// GetProjectIDByHash -
func (storage *Storage) GetProjectIDByHash(hash string) (result string, err error) {
	err = storage.DB.Table(models.DocContracts).Select("project_id").Where("hash = ?", hash).Where("project_id != ''").Limit(1).Scan(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return
}
