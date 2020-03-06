package models

import (
	"time"

	"github.com/tidwall/gjson"
)

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
	SumTxAmount       int64   `json:"sum_tx_amount,omitempty"`
	MedianConsumedGas int64   `json:"median_consumed_gas,omitempty"`
}

// Fingerprint -
type Fingerprint struct {
	Code      string `json:"code"`
	Storage   string `json:"storage"`
	Parameter string `json:"parameter"`
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
	c.SumTxAmount = hit.Get("_source.sum_tx_amount").Int()
	c.MedianConsumedGas = hit.Get("_source.median_consumed_gas").Int()

	c.FoundBy = getFoundBy(hit)
}

// ParseElasticJSON -
func (f *Fingerprint) ParseElasticJSON(hit gjson.Result) {
	f.Code = hit.Get("code").String()
	f.Parameter = hit.Get("parameter").String()
	f.Storage = hit.Get("storage").String()
}

func getFoundBy(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()

	if _, ok := keys["address"]; ok {
		return "address"
	}
	if _, ok := keys["manager"]; ok {
		return "manager"
	}
	if _, ok := keys["addredelegatess"]; ok {
		return "delegate"
	}
	if _, ok := keys["tags"]; ok {
		return "tags"
	}
	if _, ok := keys["hardcoded"]; ok {
		return "hardcoded addresses"
	}
	if _, ok := keys["annotations"]; ok {
		return "annotations"
	}
	if _, ok := keys["fail_strings"]; ok {
		return "fail strings"
	}
	if _, ok := keys["entrypoints"]; ok {
		return "entrypoints"
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
