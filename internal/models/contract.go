package models

import (
	"time"

	"github.com/tidwall/gjson"
)

var foundByCategories = []string{
	"alias",
	"tags",
	"entrypoints",
	"entrypoint",
	"fail_strings",
	"errors.with",
	"errors.id",
	"language",
	"annotations",
	"delegate",
	"hardcoded",
	"manager",
	"address",
	"hash",
	"key_hash",
	"key_strings",
	"value_strings",
	"parameter_strings",
	"storage_strings",
}

// Contract - entity for contract
type Contract struct {
	ID          string       `json:"id"`
	Network     string       `json:"network"`
	Level       int64        `json:"level"`
	Timestamp   time.Time    `json:"timestamp"`
	Balance     int64        `json:"balance"`
	Fingerprint *Fingerprint `json:"fingerprint,omitempty"`
	Language    string       `json:"language,omitempty"`

	Tags        []string `json:"tags,omitempty"`
	Hardcoded   []string `json:"hardcoded,omitempty"`
	FailStrings []string `json:"fail_strings,omitempty"`
	Primitives  []string `json:"primitives,omitempty"`
	Annotations []string `json:"annotations,omitempty"`
	Entrypoints []string `json:"entrypoints,omitempty"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty"`
	Delegate string `json:"delegate,omitempty"`

	ProjectID         string  `json:"project_id,omitempty"`
	FoundBy           string  `json:"found_by,omitempty"`
	LastAction        BCDTime `json:"last_action,omitempty"`
	TxCount           int64   `json:"tx_count,omitempty"`
	MigrationsCount   int64   `json:"migrations_count,omitempty"`
	TotalWithdrawn    int64   `json:"total_withdrawn,omitempty"`
	MedianConsumedGas int64   `json:"median_consumed_gas,omitempty"`
	Alias             string  `json:"alias,omitempty"`
	DelegateAlias     string  `json:"delegate_alias,omitempty"`
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
	c.Primitives = parseStringArray(hit, "_source.primitives")
	c.FailStrings = parseStringArray(hit, "_source.fail_strings")
	c.Entrypoints = parseStringArray(hit, "_source.entrypoints")

	f := hit.Get("_source.fingerprint")
	if f.Exists() {
		c.Fingerprint = &Fingerprint{}
		c.Fingerprint.ParseElasticJSON(f)
	}

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
	c.MedianConsumedGas = hit.Get("_source.median_consumed_gas").Int()
	c.Alias = hit.Get("_source.alias").String()

	c.FoundBy = GetFoundBy(hit)
}

// ParseElasticJSON -
func (f *Fingerprint) ParseElasticJSON(hit gjson.Result) {
	f.Code = hit.Get("code").String()
	f.Parameter = hit.Get("parameter").String()
	f.Storage = hit.Get("storage").String()
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

// GetFoundBy -
func GetFoundBy(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()

	for _, category := range foundByCategories {
		if _, ok := keys[category]; ok {
			return category
		}
	}

	for category := range keys {
		return category
	}

	return ""
}

func parseStringArray(hit gjson.Result, tag string) []string {
	res := make([]string, 0)
	for _, t := range hit.Get(tag).Array() {
		res = append(res, t.String())
	}
	return res
}
