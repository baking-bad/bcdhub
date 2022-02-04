package contract

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
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
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(network types.Network, address string) (response contract.Contract, err error) {
	var accountID int64
	if err = storage.DB.Model((*account.Account)(nil)).Column("id").Where("network = ?", network).Where("address = ?", address).Select(&accountID); err != nil {
		return
	}

	err = storage.DB.Model(&response).Where("contract.account_id = ?", accountID).Relation("Account").Relation("Manager").Relation("Delegate").Relation("Alpha").Relation("Babylon").Select()
	return
}

// GetMany -
func (storage *Storage) GetMany(network types.Network) (response []contract.Contract, err error) {
	err = storage.DB.Model(&response).Where("contract.network = ?", network).Relation("Account").Relation("Manager").Relation("Delegate").Select(&response)
	return
}

// GetRandom -
func (storage *Storage) GetRandom(networks ...types.Network) (response contract.Contract, err error) {
	query := storage.DB.Model(&response)
	if len(networks) > 0 {
		for i := range networks {
			if networks[i] != types.Empty {
				query.WhereOr("contract.network = ?", networks[i])
			}
		}
	}
	err = query.OrderExpr("random()").Relation("Account").Relation("Manager").Relation("Delegate").Relation("Alpha").Relation("Babylon").First()
	return
}

// GetByAddresses -
func (storage *Storage) GetByAddresses(addresses []contract.Address) (response []contract.Contract, err error) {
	query := storage.DB.Model((*contract.Contract)(nil))

	for i := range addresses {
		query.WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
			q.Where("contract.network = ?", addresses[i].Network).Where("account.address = ?", addresses[i].Address)
			return q, nil
		})
	}

	err = query.Relation("Account").Relation("Manager").Relation("Delegate").Select(&response)
	return
}

// GetSameContracts -
func (storage *Storage) GetSameContracts(c contract.Contract, manager string, size, offset int64) (pcr contract.SameResponse, err error) {
	limit := storage.GetPageSize(size)

	query := storage.DB.Model((*contract.Contract)(nil)).Where("account_id != ?", c.AccountID)

	if c.AlphaID > 0 {
		query.Where("alpha_id = ?", c.AlphaID)
	}
	if c.BabylonID > 0 {
		query.Where("babylon_id = ?", c.BabylonID)
	}

	var managerID int64
	if manager != "" {
		if err = storage.DB.Model((*account.Account)(nil)).Column("id").Where("network = ?", c.Network).Where("address = ?", manager).Select(&managerID); err != nil {
			return
		}
		query.Where("manager_id = ?", managerID)
	}
	query.Order("last_action desc").Relation("Account").Relation("Manager").Relation("Delegate").Relation("Alpha").Relation("Babylon").Limit(limit).Offset(int(offset))
	if err = query.Select(&pcr.Contracts); err != nil {
		return
	}

	countQuery := storage.DB.Model((*contract.Contract)(nil)).Where("account_id != ?", c.AccountID)
	if c.AlphaID > 0 {
		countQuery.Where("alpha_id = ?", c.AlphaID)
	}
	if c.BabylonID > 0 {
		countQuery.Where("babylon_id = ?", c.BabylonID)
	}
	if managerID > 0 {
		countQuery.Where("manager_id = ?", managerID)
	}
	count, err := countQuery.Count()
	if err != nil {
		return
	}
	pcr.Count = int64(count)
	return
}

// GetSimilarContracts -
func (storage *Storage) GetSimilarContracts(c contract.Contract, size, offset int64) ([]contract.Similar, int, error) {
	script := c.Alpha
	if c.BabylonID > 0 {
		script = c.Babylon
	}

	var ids []int64
	if err := storage.DB.Model((*contract.Script)(nil)).Column("id").
		Where("project_id = ?", script.ProjectID).
		Where("hash != ?", script.Hash).
		Select(&ids); err != nil {
		return nil, 0, err
	}

	if len(ids) == 0 {
		return []contract.Similar{}, 0, nil
	}

	limit := storage.GetPageSize(size)

	var contracts []contract.Contract
	if err := storage.DB.Model(&contracts).
		WhereIn("alpha_id IN (?)", ids).
		WhereIn("babylon_id IN (?)", ids).
		Relation("Account").Relation("Manager").Relation("Delegate").Relation("Alpha").Relation("Babylon").
		Order("last_action desc").
		Limit(limit).
		Offset(int(offset)).
		Select(); err != nil {
		return nil, 0, err
	}

	count, err := storage.DB.Model((*contract.Contract)(nil)).
		WhereIn("alpha_id IN (?)", ids).
		WhereIn("babylon_id IN (?)", ids).
		Count()
	if err != nil {
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
	err := storage.DB.Model((*contract.Contract)(nil)).
		Where("contract.network = ?", network).
		Where("(contract.tags & ?) > 0", tags).
		Order("contract.id desc").
		Limit(storage.GetPageSize(size)).
		Offset(int(offset)).
		Relation("Account").Relation("Manager").Relation("Delegate").
		Select(&contracts)
	if err != nil {
		return nil, 0, err
	}

	count, err := storage.DB.Model((*contract.Contract)(nil)).Where("(contract.tags & ?) > 0", tags).Where("contract.network = ?", network).Count()
	return contracts, int64(count), err
}

// Stats -
func (storage *Storage) Stats(c contract.Contract) (stats contract.Stats, err error) {
	sameCount, err := storage.DB.Model((*contract.Contract)(nil)).
		WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
			if c.AlphaID > 0 {
				q.WhereOr("alpha_id = ?", c.AlphaID)
			}
			if c.BabylonID > 0 {
				q.WhereOr("babylon_id = ?", c.BabylonID)
			}
			return q, err
		}).
		Count()
	if err != nil {
		return
	}
	stats.SameCount = int64(sameCount) - 1

	projectID := c.Alpha.ProjectID
	if c.BabylonID > 0 {
		projectID = c.Babylon.ProjectID
	}
	scriptsQuery := storage.DB.Model((*contract.Script)(nil)).Column("id").
		Where("project_id = (?)", projectID)

	similarCount, err := storage.DB.Model((*contract.Contract)(nil)).
		Where("(alpha_id is not null and alpha_id IN (?0)) OR (babylon_id is not null and babylon_id IN (?0))", scriptsQuery).
		Count()
	if err != nil {
		return
	}
	stats.SimilarCount = int64(similarCount)
	return
}

// ByHash -
func (storage *Storage) ByHash(hash string) (result contract.Script, err error) {
	err = storage.DB.Model(&result).Where("hash = ?", hash).First()
	return
}

// Script -
func (storage *Storage) Script(network types.Network, address string, symLink string) (contract.Script, error) {
	var accountID int64
	if err := storage.DB.Model((*account.Account)(nil)).Column("id").Where("network = ?", network).Where("address = ?", address).Select(&accountID); err != nil {
		return contract.Script{}, err
	}

	var c contract.Contract
	query := storage.DB.Model(&c).
		Where("contract.network = ?", network).
		Where("account_id = ?", accountID)
	switch symLink {
	case bcd.SymLinkAlpha:
		err := query.Relation("Alpha").Select()
		return c.Alpha, err
	case bcd.SymLinkBabylon:
		err := query.Relation("Babylon").Select()
		return c.Babylon, err
	}
	return c.Alpha, errors.Errorf("unknown protocol symbolic link: %s", symLink)
}

// GetScripts -
func (storage *Storage) GetScripts(limit, offset int) (scripts []contract.Script, err error) {
	err = storage.DB.Model(&scripts).Limit(limit).Offset(offset).Order("id asc").Select()
	return
}

// UpdateProjectID -
func (storage *Storage) UpdateProjectID(scripts []contract.Script) error {
	_, err := storage.DB.Model(&scripts).Set("project_id = _data.project_id").WherePK().Update()
	return err
}

// Code -
func (storage *Storage) Code(id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.Model((*contract.Script)(nil)).Where("id = ?", id).Column("code").Select(&data)
	return data, err
}

// Parameter -
func (storage *Storage) Parameter(id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.Model((*contract.Script)(nil)).Where("id = ?", id).Column("parameter").Select(&data)
	return data, err
}

// Storage -
func (storage *Storage) Storage(id int64) ([]byte, error) {
	var data []byte
	err := storage.DB.Model((*contract.Script)(nil)).Where("id = ?", id).Column("storage").Select(&data)
	return data, err
}

// ScriptPart -
func (storage *Storage) ScriptPart(network types.Network, address string, symLink, part string) ([]byte, error) {
	var accountID int64
	if err := storage.DB.Model((*account.Account)(nil)).Column("id").Where("network = ?", network).Where("address = ?", address).Select(&accountID); err != nil {
		return nil, err
	}

	query := storage.DB.Model((*contract.Contract)(nil)).
		Where("contract.network = ?", network).
		Where("account_id = ?", accountID)

	switch symLink {
	case "alpha":
		switch part {
		case "parameter":
			query.Column("alpha.parameter").Relation("Alpha._")
		case "code":
			query.Column("alpha.code").Relation("Alpha._")
		case "storage":
			query.Column("alpha.storage").Relation("Alpha._")
		default:
			return nil, errors.Errorf("unknown script part name: %s", part)
		}
	case "babylon":
		switch part {
		case "parameter":
			query.Column("babylon.parameter").Relation("Babylon._")
		case "code":
			query.Column("babylon.code").Relation("Babylon._")
		case "storage":
			query.Column("babylon.storage").Relation("Babylon._")
		default:
			return nil, errors.Errorf("unknown script part name: %s", part)
		}
	default:
		return nil, errors.Errorf("unknown protocol symbolic link: %s", symLink)
	}
	var data []byte
	err := query.Select(pg.Scan(&data))
	return data, err
}
