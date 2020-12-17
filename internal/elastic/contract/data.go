package contract

import "github.com/baking-bad/bcdhub/internal/elastic/core"

type getContractMigrationStatsResponse struct {
	Agg struct {
		MigrationsCount core.IntValue `json:"migrations_count"`
	} `json:"aggregations"`
}

type operationAddresses struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type getDiffTasksResponse struct {
	Agg struct {
		Projects struct {
			Buckets []struct {
				core.Bucket
				Last struct {
					Hits core.HitsArray `json:"hits"`
				} `json:"last"`
				ByHash struct {
					Buckets []struct {
						core.Bucket
						Last struct {
							Hits core.HitsArray `json:"hits"`
						} `json:"last"`
					} `json:"buckets"`
				} `json:"by_hash"`
			} `json:"buckets"`
		} `json:"by_project"`
	} `json:"aggregations"`
}

type getProjectsResponse struct {
	Agg struct {
		Projects struct {
			Buckets []struct {
				core.Bucket
				Last struct {
					Hits core.HitsArray `json:"hits"`
				} `json:"last"`
			} `json:"buckets"`
		} `json:"projects"`
	} `json:"aggregations"`
}

type getDAppStatsResponse struct {
	Aggs struct {
		Users  core.FloatValue `json:"users"`
		Calls  core.FloatValue `json:"calls"`
		Volume core.FloatValue `json:"volume"`
	} `json:"aggregations"`
}

type recalcContractStatsResponse struct {
	Aggs struct {
		TxCount        core.IntValue `json:"tx_count"`
		Balance        core.IntValue `json:"balance"`
		LastAction     core.IntValue `json:"last_action"`
		TotalWithdrawn core.IntValue `json:"total_withdrawn"`
	} `json:"aggregations"`
}
