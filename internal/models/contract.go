package models

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Contract - entity for contract
type Contract struct {
	ID        string    `json:"id"`
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Balance   int64     `json:"balance"`
	Language  string    `json:"language,omitempty"`

	Hash        string       `json:"hash"`
	Fingerprint *Fingerprint `json:"fingerprint,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	Hardcoded   []string     `json:"hardcoded,omitempty"`
	FailStrings []string     `json:"fail_strings,omitempty"`
	Annotations []string     `json:"annotations,omitempty"`
	Entrypoints []string     `json:"entrypoints,omitempty"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty"`
	Delegate string `json:"delegate,omitempty"`

	ProjectID       string  `json:"project_id,omitempty"`
	FoundBy         string  `json:"found_by,omitempty"`
	LastAction      BCDTime `json:"last_action,omitempty"`
	TxCount         int64   `json:"tx_count,omitempty"`
	MigrationsCount int64   `json:"migrations_count,omitempty"`
	TotalWithdrawn  int64   `json:"total_withdrawn,omitempty"`
	Alias           string  `json:"alias,omitempty"`
	DelegateAlias   string  `json:"delegate_alias,omitempty"`
}

// GetID -
func (c *Contract) GetID() string {
	return c.ID
}

// GetIndex -
func (c *Contract) GetIndex() string {
	return "contract"
}

// GetQueue -
func (c *Contract) GetQueue() string {
	return "contracts"
}

// ParseElasticJSON -
func (c *Contract) ParseElasticJSON(hit gjson.Result) {
	c.ID = hit.Get("_id").String()
	c.Network = hit.Get("_source.network").String()
	c.Level = hit.Get("_source.level").Int()
	c.Timestamp = hit.Get("_source.timestamp").Time().UTC()
	c.Balance = hit.Get("_source.balance").Int()
	c.Language = hit.Get("_source.language").String()

	c.Tags = parseStringArray(hit, "_source.tags")
	c.Hardcoded = parseStringArray(hit, "_source.hardcoded")
	c.Annotations = parseStringArray(hit, "_source.annotations")
	c.FailStrings = parseStringArray(hit, "_source.fail_strings")
	c.Entrypoints = parseStringArray(hit, "_source.entrypoints")

	f := hit.Get("_source.fingerprint")
	if f.Exists() {
		c.Fingerprint = &Fingerprint{}
		c.Fingerprint.ParseElasticJSON(f)
	}

	c.Hash = hit.Get("_source.hash").String()
	c.Address = hit.Get("_source.address").String()
	c.Manager = hit.Get("_source.manager").String()
	c.Delegate = hit.Get("_source.delegate").String()

	c.ProjectID = hit.Get("_source.project_id").String()

	c.LastAction = BCDTime{
		Time: hit.Get("_source.last_action").Time().UTC(),
	}

	c.TxCount = hit.Get("_source.tx_count").Int()
	c.MigrationsCount = hit.Get("_source.migrations_count").Int()
	c.TotalWithdrawn = hit.Get("_source.total_withdrawn").Int()
	c.Alias = hit.Get("_source.alias").String()

	c.FoundBy = c.FoundByName(hit)
}

// GetScores -
func (c *Contract) GetScores(search string) []string {
	if helpers.IsAddress(search) {
		return []string{
			"address^10",
			"alias^9",
			"tags^9",
			"entrypoints^8",
			"fail_strings^6",
			"language^4",
			"annotations^3",
			"delegate^2",
			"hardcoded^2",
			"manager",
		}
	}
	return []string{
		"alias^10",
		"tags^9",
		"entrypoints^8",
		"fail_strings^6",
		"language^4",
		"annotations^3",
		"delegate^2",
		"hardcoded^2",
		"manager",
		"address",
	}
}

// FoundByName -
func (c *Contract) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := c.GetScores("")
	return getFoundBy(keys, categories)
}

// IsFA12 - checks contract realizes fa12 interface
func (c *Contract) IsFA12() bool {
	for i := range c.Tags {
		if c.Tags[i] == "fa12" {
			return true
		}
	}
	return false
}

// Fingerprint -
type Fingerprint struct {
	Code      string `json:"code"`
	Storage   string `json:"storage"`
	Parameter string `json:"parameter"`
}

// Compare -
func (f *Fingerprint) Compare(second *Fingerprint) bool {
	return f.Code == second.Code && f.Parameter == second.Parameter && f.Storage == second.Storage
}

// ParseElasticJSON -
func (f *Fingerprint) ParseElasticJSON(hit gjson.Result) {
	f.Code = hit.Get("code").String()
	f.Parameter = hit.Get("parameter").String()
	f.Storage = hit.Get("storage").String()
}

// BCDTime -
type BCDTime struct {
	time.Time
}

// MarshalJSON -
func (t BCDTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return t.Time.MarshalJSON()
}
