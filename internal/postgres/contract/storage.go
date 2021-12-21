package contract

import (
	"encoding/hex"
	"math/rand"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
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
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(network types.Network, address string) (response contract.Contract, err error) {
	query := storage.DB.Model(&response)
	core.NetworkAndAddress(network, address)(query)
	err = query.First()
	return
}

// GetMany -
func (storage *Storage) GetMany(by map[string]interface{}) (response []contract.Contract, err error) {
	query := storage.DB.Model().Table(models.DocContracts)
	for column, value := range by {
		query.Where("? = ?", pg.Ident(column), value)
	}
	err = query.Select(&response)
	return
}

// GetRandom -
func (storage *Storage) GetRandom(network types.Network) (response contract.Contract, err error) {
	queryCount := storage.DB.Model(&response).Where("tx_count > 2")
	if network != types.Empty {
		queryCount.Where("network = ?", network)
	}
	count, err := queryCount.Count()
	if err != nil {
		return
	}

	if count == 0 {
		return response, pg.ErrNoRows
	}

	query := storage.DB.Model(&response).Where("tx_count > 2")
	if network != types.Empty {
		query.Where("network = ?", network)
	}
	err = query.Offset(rand.Intn(count)).First()
	return
}

// GetByAddresses -
func (storage *Storage) GetByAddresses(addresses []contract.Address) (response []contract.Contract, err error) {
	query := storage.DB.Model().Table(models.DocContracts)

	for i := range addresses {
		query.WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
			core.NetworkAndAddress(addresses[i].Network, addresses[i].Address)(q)
			return q, nil
		})
	}

	err = query.Select(&response)
	return
}

// GetProjectsLastContract -
func (storage *Storage) GetProjectsLastContract(c contract.Contract, size, offset int64) (response []*contract.Contract, err error) {
	if c.FingerprintCode == nil || c.FingerprintParameter == nil || c.FingerprintStorage == nil {
		return nil, nil
	}

	code := hex.EncodeToString(c.FingerprintCode)
	params := hex.EncodeToString(c.FingerprintParameter)
	s := hex.EncodeToString(c.FingerprintStorage)

	limit := storage.GetPageSize(size)

	query := storage.DB.Model().Table(models.DocContracts).
		ColumnExpr("MAX(id) as id").
		Where("project_id is not null").
		Where("encode(fingerprint_code, 'hex') = ?", code).
		Where("encode(fingerprint_parameter, 'hex') = ?", params).
		Where("encode(fingerprint_storage, 'hex') = ?", s)

	if c.Manager.Valid {
		query.WhereOr("manager = ?", c.Manager.String())
	}

	query.Group("project_id").
		Limit(limit).
		Offset(int(offset)).
		Order("id desc")

	err = storage.DB.Model().Table(models.DocContracts).Where("id IN (?)", query).Select(&response)
	return
}

// GetSameContracts -
func (storage *Storage) GetSameContracts(c contract.Contract, manager string, size, offset int64) (pcr contract.SameResponse, err error) {
	if c.FingerprintCode == nil || c.FingerprintParameter == nil || c.FingerprintStorage == nil {
		return pcr, errors.Wrap(consts.ErrInvalidFingerprint, c.Address)
	}

	limit := storage.GetPageSize(size)

	query := storage.DB.Model().Table(models.DocContracts).Where("hash = ?", c.Hash).Where("address != ?", c.Address)
	if manager != "" {
		query.Where("manager = ?", manager)
	}
	query.Order("last_action desc").Limit(limit).Offset(int(offset))
	if err = query.Select(&pcr.Contracts); err != nil {
		return
	}

	countQuery := storage.DB.Model().Table(models.DocContracts).Where("hash = ?", c.Hash).Where("address != ?", c.Address)
	if manager != "" {
		countQuery.Where("manager = ?", manager)
	}
	count, err := countQuery.Order("last_action desc").Count()
	if err != nil {
		return
	}
	pcr.Count = int64(count)
	return
}

// GetSimilarContracts -
func (storage *Storage) GetSimilarContracts(c contract.Contract, size, offset int64) ([]contract.Similar, int, error) {
	if !c.ProjectID.Valid {
		return nil, 0, nil
	}

	limit := storage.GetPageSize(size)

	subQuery := storage.DB.Model((*contract.Contract)(nil)).
		ColumnExpr("MAX(id) as id").
		Where("project_id = ?", c.ProjectID).
		Where("hash != ?", c.Hash).
		Group("hash")

	var contracts []contract.Contract
	if err := storage.DB.Model((*contract.Contract)(nil)).
		Where("id IN (?)", subQuery).
		Order("last_action desc").
		Limit(limit).
		Offset(int(offset)).
		Select(&contracts); err != nil {
		return nil, 0, err
	}

	var count int
	if err := storage.DB.Model((*contract.Contract)(nil)).
		ColumnExpr("count(distinct hash)").
		Where("project_id = ?", c.ProjectID).
		Where("hash != ?", c.Hash).
		Group("hash").
		Select(&count); err != nil {
		return nil, 0, err
	}

	pcr := make([]contract.Similar, len(contracts))
	for i := range contracts {
		pcr[i].Contract = &contracts[i]
	}
	return pcr, count, nil
}

// GetTokens -
func (storage *Storage) GetTokens(network types.Network, tokenInterface string, offset, size int64) ([]contract.Contract, int64, error) {
	tags := types.FA12Tag | types.FA1Tag | types.FA2Tag
	if tokenInterface == "fa1-2" || tokenInterface == "fa1" || tokenInterface == "fa2" {
		tags = types.NewTags([]string{tokenInterface})
	}

	var contracts []contract.Contract
	query := storage.DB.Model((*contract.Contract)(nil))
	core.Network(network)(query)

	err := query.Where("(tags & ?) > 0", tags).
		Order("id desc").
		Limit(storage.GetPageSize(size)).
		Offset(int(offset)).
		Select(&contracts)
	if err != nil {
		return nil, 0, err
	}

	countQuery := storage.DB.Model().Table(models.DocContracts).Where("(tags & ?) > 0", tags)
	core.Network(network)(countQuery)
	count, err := countQuery.Count()
	return contracts, int64(count), err
}

// Stats -
func (storage *Storage) Stats(c contract.Contract) (stats contract.Stats, err error) {
	if !c.ProjectID.Valid {
		return
	}
	sameCount, err := storage.DB.Model().Table(models.DocContracts).
		Where("hash = ?", c.Hash).
		Where("address != ?", c.Address).
		Count()
	if err != nil {
		return
	}
	stats.SameCount = int64(sameCount)

	if err = storage.DB.Model((*contract.Contract)(nil)).
		ColumnExpr("count(distinct hash)").
		Where("project_id = ?", c.ProjectID).
		Where("hash != ?", c.Hash).
		Group("hash").
		Select(&stats.SimilarCount); err != nil {
		return
	}

	return
}

// GetProjectIDByHash -
func (storage *Storage) GetProjectIDByHash(hash string) (result string, err error) {
	err = storage.DB.Model().Table(models.DocContracts).Column("project_id").Where("hash = ?", hash).Where("project_id is not null").Limit(1).Select(&result)
	if errors.Is(err, pg.ErrNoRows) {
		return "", nil
	}
	return
}
