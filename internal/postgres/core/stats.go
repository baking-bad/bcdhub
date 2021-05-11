package core

import "github.com/baking-bad/bcdhub/internal/models"

// GetNetworkCountStats -
func (p *Postgres) GetNetworkCountStats(network string) (map[string]int64, error) {
	var contractsCount int64
	if err := p.DB.Table(models.DocContracts).Where("network = ?", network).Count(&contractsCount).Error; err != nil {
		return nil, err
	}
	var operationsCount int64
	if err := p.DB.Table(models.DocOperations).Where("network = ?", network).Count(&operationsCount).Error; err != nil {
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

type stats struct {
	Network   string
	Value     uint64
	StatsType string
}

// GetStats -
func (p *Postgres) GetStats(network string) (map[string]*models.NetworkStats, error) {
	var s []stats
	if err := p.DB.Table("head_stats").Find(&s).Error; err != nil {
		return nil, err
	}

	result := make(map[string]*models.NetworkStats)
	for i := range s {
		if _, ok := result[s[i].Network]; !ok {
			result[s[i].Network] = new(models.NetworkStats)
		}

		val := result[s[i].Network]

		switch s[i].StatsType {
		case "calls_count":
			val.CallsCount = s[i].Value
		case "fa_count":
			val.FACount = s[i].Value
		case "unique_contracts_count":
			val.UniqueContractsCount = s[i].Value
		case "contracts_count":
			val.ContractsCount = s[i].Value
		}
	}
	return result, nil
}
