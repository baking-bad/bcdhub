package elastic

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/models"
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

// LightContract -
type LightContract struct {
	Address  string    `json:"address"`
	Network  string    `json:"network"`
	Deployed time.Time `json:"deploy_time"`
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

// SimilarContract -
type SimilarContract struct {
	*models.Contract
	Count uint64 `json:"count"`
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

// DiffTask -
type DiffTask struct {
	Network1 string
	Address1 string
	Network2 string
	Address2 string
}

// ContractCountStats -
type ContractCountStats struct {
	Total          int64
	SameCount      int64
	TotalWithdrawn int64
	Balance        int64
}

// SubscriptionRequest -
type SubscriptionRequest struct {
	Address         string
	Network         string
	Alias           string
	Hash            string
	ProjectID       string
	WithSame        bool
	WithSimilar     bool
	WithMempool     bool
	WithMigrations  bool
	WithErrors      bool
	WithCalls       bool
	WithDeployments bool
}

// EventContract -
type EventContract struct {
	Network   string    `json:"network"`
	Address   string    `json:"address"`
	Hash      string    `json:"hash"`
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ParseElasticJSON -
func (m *EventContract) ParseElasticJSON(resp gjson.Result) {
	m.Network = resp.Get("_source.network").String()
	m.Address = resp.Get("_source.address").String()
	m.Hash = resp.Get("_source.hash").String()
	m.ProjectID = resp.Get("_source.project_id").String()
	m.Timestamp = resp.Get("_source.timestamp").Time().UTC()
}

// EventType -
const (
	EventTypeError     = "error"
	EventTypeMigration = "migration"
	EventTypeCall      = "call"
	EventTypeInvoke    = "invoke"
	EventTypeDeploy    = "deploy"
	EventTypeSame      = "same"
	EventTypeSimilar   = "similar"
	EventTypeMempool   = "mempool"
)

// Event -
type Event struct {
	Type    string      `json:"type"`
	Address string      `json:"address"`
	Network string      `json:"network"`
	Alias   string      `json:"alias"`
	Body    interface{} `json:"body,omitempty"`
}

// EventOperation -
type EventOperation struct {
	Network          string    `json:"network"`
	Hash             string    `json:"hash"`
	Internal         bool      `json:"internal"`
	Status           string    `json:"status"`
	Timestamp        time.Time `json:"timestamp"`
	Kind             string    `json:"kind"`
	Fee              int64     `json:"fee,omitempty"`
	Amount           int64     `json:"amount,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	Source           string    `json:"source"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	Destination      string    `json:"destination,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`
	Delegate         string    `json:"delegate,omitempty"`
	DelegateAlias    string    `json:"delegate_alias,omitempty"`

	Result *models.OperationResult `json:"result,omitempty"`
	Errors []cerrors.IError        `json:"errors,omitempty"`
	Burned int64                   `json:"burned,omitempty"`
}

// ParseElasticJSON -
func (o *EventOperation) ParseElasticJSON(resp gjson.Result) {
	o.Hash = resp.Get("_source.hash").String()
	o.Internal = resp.Get("_source.internal").Bool()
	o.Network = resp.Get("_source.network").String()
	o.Timestamp = resp.Get("_source.timestamp").Time().UTC()
	o.Status = resp.Get("_source.status").String()
	o.Kind = resp.Get("_source.kind").String()
	o.Source = resp.Get("_source.source").String()
	o.Fee = resp.Get("_source.fee").Int()
	o.Amount = resp.Get("_source.amount").Int()
	o.Destination = resp.Get("_source.destination").String()
	o.Delegate = resp.Get("_source.delegate").String()
	o.Entrypoint = resp.Get("_source.entrypoint").String()
	o.SourceAlias = resp.Get("_source.source_alias").String()
	o.DestinationAlias = resp.Get("_source.destination_alias").String()
	o.Burned = resp.Get("_source.burned").Int()

	var opResult models.OperationResult
	opResult.ParseElasticJSON(resp.Get("_source.result"))
	o.Result = &opResult

	err := resp.Get("_source.errors")
	o.Errors = cerrors.ParseArray(err)
}

// EventMigration -
type EventMigration struct {
	Network      string    `json:"network"`
	Protocol     string    `json:"protocol"`
	PrevProtocol string    `json:"prev_protocol,omitempty"`
	Hash         string    `json:"hash,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Level        int64     `json:"level"`
	Address      string    `json:"address"`
	Kind         string    `json:"kind"`
}

// ParseElasticJSON -
func (m *EventMigration) ParseElasticJSON(resp gjson.Result) {
	m.Protocol = resp.Get("_source.protocol").String()
	m.PrevProtocol = resp.Get("_source.prev_protocol").String()
	m.Hash = resp.Get("_source.hash").String()
	m.Network = resp.Get("_source.network").String()
	m.Timestamp = resp.Get("_source.timestamp").Time().UTC()
	m.Level = resp.Get("_source.level").Int()
	m.Address = resp.Get("_source.address").String()
	m.Kind = resp.Get("_source.kind").String()
}

// TokenMethodUsageStats -
type TokenMethodUsageStats struct {
	Count       int64
	ConsumedGas int64
}

// TokenUsageStats -
type TokenUsageStats map[string]TokenMethodUsageStats

// DAppStats -
type DAppStats struct {
	Users  int64 `json:"users"`
	Calls  int64 `json:"txs"`
	Volume int64 `json:"volume"`
}

// ParseElasticJSON -
func (stats *DAppStats) ParseElasticJSON(hit gjson.Result) {
	stats.Calls = hit.Get("calls.value").Int()
	stats.Users = hit.Get("users.value").Int()
	stats.Volume = hit.Get("volume.value").Int()
}

// TransfersResponse -
type TransfersResponse struct {
	Transfers []models.Transfer `json:"transfers"`
	Total     int64             `json:"total"`
	LastID    string            `json:"last_id"`
}
