package elastic

import (
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/goombaio/namegenerator"
	"github.com/tidwall/gjson"
)

// SearchResult -
type SearchResult struct {
	Count int64        `json:"count"`
	Time  int64        `json:"time"`
	Items []SearchItem `json:"items"`
}

// SearchItem -
type SearchItem struct {
	Type       string              `json:"type"`
	Value      string              `json:"value"`
	Group      *Group              `json:"group,omitempty"`
	Body       interface{}         `json:"body"`
	Highlights map[string][]string `json:"highlights,omitempty"`
}

// Group -
type Group struct {
	Count int64 `json:"count"`
	Top   []Top `json:"top"`
}

// Top -
type Top struct {
	Network string `json:"network"`
	Key     string `json:"key"`
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
	stats.Name = stats.getName(data)
	stats.Last = LightContract{
		Address:  data.Get("last.hits.hits.0._source.address").String(),
		Network:  data.Get("last.hits.hits.0._source.network").String(),
		Deployed: data.Get("last.hits.hits.0._source.timestamp").Time().UTC(),
	}
}

func (stats *ProjectStats) getName(data gjson.Result) string {
	if data.Get("alias").String() != "" {
		return data.Get("alias").String()
	}
	s := data.Get("key").String()[:8]
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
	ConsumedGasDiff float64 `json:"consumed_gas_diff"`
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

// BigMapDiff -
type BigMapDiff struct {
	Ptr         int64     `json:"ptr,omitempty"`
	BinPath     string    `json:"bin_path"`
	Key         string    `json:"key"`
	KeyHash     string    `json:"key_hash"`
	Value       string    `json:"value"`
	OperationID string    `json:"operation_id"`
	Level       int64     `json:"level"`
	Address     string    `json:"address"`
	Network     string    `json:"network"`
	Timestamp   time.Time `json:"timestamp"`
	Protocol    string    `json:"protocol"`

	Count int64 `json:"count"`
}

// ParseElasticJSON -
func (b *BigMapDiff) ParseElasticJSON(hit gjson.Result) {
	b.Ptr = hit.Get("_source.ptr").Int()
	b.BinPath = hit.Get("_source.bin_path").String()
	b.Key = hit.Get("_source.key").String()
	b.KeyHash = hit.Get("_source.key_hash").String()
	b.Value = hit.Get("_source.value").String()
	b.OperationID = hit.Get("_source.operation_id").String()
	b.Level = hit.Get("_source.level").Int()
	b.Address = hit.Get("_source.address").String()
	b.Network = hit.Get("_source.network").String()
	b.Timestamp = hit.Get("_source.timestamp").Time()
	b.Protocol = hit.Get("_source.protocol").String()
}

type contractPair struct {
	Address string
	Network string
}

// ParseElasticJSONArray -
func (c *contractPair) ParseElasticJSONArray(hit gjson.Result) {
	c.Address = hit.Get("0").String()
	c.Network = hit.Get("1").String()
}

// TimelineItem -
type TimelineItem struct {
	Network          string    `json:"network"`
	Hash             string    `json:"hash"`
	Status           string    `json:"status"`
	Timestamp        time.Time `json:"timestamp"`
	Kind             string    `json:"kind"`
	Source           string    `json:"source"`
	Amount           int64     `json:"amount,omitempty"`
	Level            int64     `json:"level,omitempty"`
	Destination      string    `json:"destination,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`
}

// ParseJSONOperation -
func (t *TimelineItem) ParseJSONOperation(hit gjson.Result) {
	t.Network = hit.Get("_source.network").String()
	t.Hash = hit.Get("_source.hash").String()
	t.Status = hit.Get("_source.status").String()
	t.Timestamp = hit.Get("_source.timestamp").Time()
	t.Kind = hit.Get("_source.kind").String()
	t.Source = hit.Get("_source.source").String()
	t.Destination = hit.Get("_source.destination").String()
	t.Entrypoint = hit.Get("_source.entrypoint").String()
	t.Amount = hit.Get("_source.amount").Int()
	t.Level = hit.Get("_source.level").Int()
	t.SourceAlias = hit.Get("_source.source_alias").String()
	t.DestinationAlias = hit.Get("_source.destination_alias").String()
}

// ParseJSONMigration -
func (t *TimelineItem) ParseJSONMigration(hit gjson.Result) {
	t.Network = hit.Get("_source.network").String()
	t.Hash = hit.Get("_source.hash").String()
	t.Status = "applied"
	t.Timestamp = hit.Get("_source.timestamp").Time()
	t.Kind = DocMigrations
	t.Source = hit.Get("_source.address").String()
	t.Level = hit.Get("_source.level").Int()
	t.SourceAlias = hit.Get("_source.source_alias").String()
}

// SearchBigMapDiff -
type SearchBigMapDiff struct {
	Ptr       int64     `json:"ptr"`
	Key       string    `json:"key"`
	KeyHash   string    `json:"key_hash"`
	Value     string    `json:"value"`
	Level     int64     `json:"level"`
	Address   string    `json:"address"`
	Network   string    `json:"network"`
	Timestamp time.Time `json:"timestamp"`
	FoundBy   string    `json:"found_by"`
}

// ParseElasticJSON -
func (b *SearchBigMapDiff) ParseElasticJSON(hit gjson.Result) {
	key := gjson.Parse(hit.Get("_source.key").String())
	b.Ptr = hit.Get("_source.ptr").Int()
	b.Key = stringer.Stringify(key)
	b.KeyHash = hit.Get("_source.key_hash").String()
	b.Value = hit.Get("_source.value").String()
	b.Level = hit.Get("_source.level").Int()
	b.Address = hit.Get("_source.address").String()
	b.Network = hit.Get("_source.network").String()
	b.Timestamp = hit.Get("_source.timestamp").Time()
	b.FoundBy = models.GetFoundBy(hit)
}

// ContractStats -
type ContractStats struct {
	TxCount        int64     `json:"tx_count"`
	LastAction     time.Time `json:"last_action"`
	Balance        int64     `json:"balance"`
	TotalWithdrawn int64     `json:"total_withdrawn"`
}

// ParseElasticJSON -
func (stats *ContractStats) ParseElasticJSON(hit gjson.Result) {
	stats.TxCount = hit.Get("tx_count.value").Int()
	stats.LastAction = time.Unix(0, hit.Get("last_action.value").Int()*1000000).UTC()
	stats.Balance = hit.Get("balance.value").Int()
	stats.TotalWithdrawn = hit.Get("total_withdrawn.value").Int()
}

// ContractMigrationsStats -
type ContractMigrationsStats struct {
	MigrationsCount int64 `json:"migrations_count"`
}

// ParseElasticJSON -
func (stats *ContractMigrationsStats) ParseElasticJSON(hit gjson.Result) {
	stats.MigrationsCount = hit.Get("migrations_count.value").Int()
}

// NetworkCountStats -
type NetworkCountStats struct {
	Contracts  int64 `json:"contracts"`
	Operations int64 `json:"operations"`
}
