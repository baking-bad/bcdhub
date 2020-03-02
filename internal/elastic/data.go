package elastic

import (
	"strconv"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/goombaio/namegenerator"
	"github.com/tidwall/gjson"
)

// SearchResult -
type SearchResult struct {
	Count     int64             `json:"count"`
	Time      int64             `json:"time"`
	Contracts []models.Contract `json:"contracts"`
}

// ContractStats -
type ContractStats struct {
	TxCount           int64
	SumTxAmount       int64
	MedianConsumedGas int64
	LastAction        time.Time
}

func (c *ContractStats) parse(data gjson.Result) {
	c.LastAction = data.Get("last_action.value_as_string").Time().UTC()
	c.TxCount = data.Get("tx_count.value").Int()
	c.SumTxAmount = data.Get("sum_tx_amount.value").Int()
}

// ProjectStats -
type ProjectStats struct {
	TxCount        int64         `json:"tx_count"`
	LastAction     time.Time     `json:"last_action"`
	LastDeploy     time.Time     `json:"last_deploy"`
	FirstDeploy    time.Time     `json:"first_deploy"`
	VersionsCount  int64         `json:"versions_count"`
	ContractsCount int64         `json:"contracts_count"`
	Language       string        `json:"language"`
	Name           string        `json:"name"`
	Last           LightContract `json:"last"`
}

// LightContract -
type LightContract struct {
	Address  string    `json:"address"`
	Network  string    `json:"network"`
	Deployed time.Time `json:"deploy_time"`
}

func (stats *ProjectStats) parse(data gjson.Result) {
	stats.FirstDeploy = time.Unix(0, data.Get("first_deploy_date.value").Int()*1000000).UTC()
	stats.LastAction = time.Unix(0, data.Get("last_action_date.value").Int()*1000000).UTC()
	stats.TxCount = data.Get("tx_count.value").Int()
	stats.VersionsCount = data.Get("count.value").Int()
	stats.ContractsCount = data.Get("doc_count").Int()
	stats.Language = data.Get("language.buckets.0.key").String()
	stats.Name = stats.getName(data.Get("key").String())
	stats.Last = LightContract{
		Address:  data.Get("last.hits.hits.0._source.address").String(),
		Network:  data.Get("last.hits.hits.0._source.network").String(),
		Deployed: data.Get("last.hits.hits.0._source.timestamp").Time().UTC(),
	}
}

func (stats *ProjectStats) getName(id string) string {
	s := id[:8]
	n, _ := strconv.ParseInt(s, 16, 64)
	nameGenerator := namegenerator.NewNameGenerator(n)
	name := nameGenerator.Generate()
	return name
}

// SimilarContract -
type SimilarContract struct {
	*models.Contract
	Count           int64   `json:"count"`
	Diff            string  `json:"diff,omitempty"`
	Added           int64   `json:"added,omitempty"`
	Removed         int64   `json:"removed,omitempty"`
	ConsumedGasDiff float64 `json:"consumed_gas_diff,omitempty"`
}

// PageableOperations -
type PageableOperations struct {
	Operations []models.Operation `json:"operations"`
	LastID     string             `json:"last_id"`
}

// SameContractsResponse -
type SameContractsResponse struct {
	Count     uint64            `json:"count"`
	Contracts []models.Contract `json:"contracts"`
}
