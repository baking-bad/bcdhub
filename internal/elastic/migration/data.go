package migration

import "github.com/baking-bad/bcdhub/internal/elastic/core"

type getContractMigrationCountResponse struct {
	Agg struct {
		MigrationsCount core.IntValue `json:"migrations_count"`
	} `json:"aggregations"`
}
