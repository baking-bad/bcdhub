package core

import "github.com/baking-bad/bcdhub/internal/models"

// GetNetworkCountStats -
func (p *Postgres) GetNetworkCountStats(network string) (map[string]int64, error) {
	var contractsCount int64
	if err := p.DB.Table(models.DocContracts).Count(&contractsCount).Error; err != nil {
		return nil, err
	}
	var operationsCount int64
	if err := p.DB.Table(models.DocOperations).Count(&operationsCount).Error; err != nil {
		return nil, err
	}
	return map[string]int64{
		models.DocContracts:  contractsCount,
		models.DocOperations: operationsCount,
	}, nil
}

type networkCount struct {
	Network string
	Count   int64
}

// GetCallsCountByNetwork -
func (p *Postgres) GetCallsCountByNetwork(network string) (map[string]int64, error) {
	query := p.DB.Table(models.DocOperations).Select("network, COUNT(*) as count").Where("entrypoint != ''")
	if network != "" {
		query.Where("network = ?", network)
	}
	query.Group("network")

	var stats []networkCount
	if err := query.Find(&stats).Error; err != nil {
		return nil, err
	}

	res := make(map[string]int64)
	for _, s := range stats {
		res[s.Network] = s.Count
	}
	return res, nil
}

type networkContractStats struct {
	Network string
	Total   int64
	Unique  int64
}

// GetContractStatsByNetwork -
func (p *Postgres) GetContractStatsByNetwork(network string) (map[string]models.ContractCountStats, error) {
	var stats []networkContractStats

	query := p.DB.Table(models.DocContracts).
		Select("network, count(*) as total, count(distinct(hash)) as unique")

	if network != "" {
		query.Where("network = ?", network)
	}

	query.
		Group("network").
		Find(&stats)

	if query.Error != nil {
		return nil, query.Error
	}

	result := make(map[string]models.ContractCountStats)
	for _, s := range stats {
		result[s.Network] = models.ContractCountStats{
			Total:     s.Total,
			SameCount: s.Unique,
		}
	}

	return result, nil
}

// GetFACountByNetwork -
func (p *Postgres) GetFACountByNetwork(network string) (map[string]int64, error) {
	var stats []networkCount
	query := p.DB.Table(models.DocContracts).
		Select("count(*) as count, network").
		Where("ARRAY['fa1', 'fa1-2', 'fa2'] && tags")

	if network != "" {
		query.Where("network = ?", network)
	}

	query.Group("network").Find(&stats)

	if query.Error != nil {
		return nil, query.Error
	}

	result := make(map[string]int64)
	for _, s := range stats {
		result[s.Network] = s.Count
	}

	return result, nil
}

// GetLanguagesForNetwork -
func (p *Postgres) GetLanguagesForNetwork(network string) (map[string]int64, error) {
	var stats []networkCount

	query := p.DB.Table(models.DocContracts).
		Select("language as network, count(*) as count").
		Where("network = ?", network).
		Group("language").
		Find(&stats)

	if query.Error != nil {
		return nil, query.Error
	}

	result := make(map[string]int64)
	for _, s := range stats {
		result[s.Network] = s.Count
	}

	return result, nil
}
