package elastic

import (
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
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
	TxCount     int64
	SumTxAmount int64
	LastAction  time.Time
}

func (c *ContractStats) parse(data gjson.Result) {
	c.LastAction = data.Get("last_action.value_as_string").Time().UTC()
	c.TxCount = data.Get("tx_count.value").Int()
	c.SumTxAmount = data.Get("sum_tx_amount.value").Int()
}

// ProjectStats -
type ProjectStats struct {
	TxCount       int64     `json:"tx_count"`
	LastAction    time.Time `json:"last_action"`
	LastDeploy    time.Time `json:"last_deploy"`
	FirstDeploy   time.Time `json:"first_deploy"`
	VersionsCount int64     `json:"versions_count"`
	Language      string    `json:"language"`
	Name          string    `json:"name"`
}

func (stats *ProjectStats) parse(data gjson.Result) {
	stats.FirstDeploy = time.Unix(0, data.Get("first_deploy_date.value").Int()*1000000).UTC()
	stats.LastAction = time.Unix(0, data.Get("last_action_date.value").Int()*1000000).UTC()
	stats.LastDeploy = time.Unix(0, data.Get("last_deploy_date.value").Int()*1000000).UTC()
	stats.TxCount = data.Get("tx_count.value").Int()
	stats.VersionsCount = data.Get("count.value").Int()
	stats.Language = data.Get("language.buckets.0.key").String()
	stats.Name = data.Get("key").String()
}
