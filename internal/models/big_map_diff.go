package models

import (
	"time"

	"github.com/tidwall/gjson"
)

// BigMapDiff -
type BigMapDiff struct {
	ID           string      `json:"-"`
	Ptr          int64       `json:"ptr,omitempty"`
	BinPath      string      `json:"bin_path"`
	Key          interface{} `json:"key"`
	KeyHash      string      `json:"key_hash"`
	KeyStrings   []string    `json:"key_strings"`
	Value        string      `json:"value"`
	ValueStrings []string    `json:"value_strings"`
	OperationID  string      `json:"operation_id"`
	Level        int64       `json:"level"`
	Address      string      `json:"address"`
	Network      string      `json:"network"`
	IndexedTime  int64       `json:"indexed_time"`
	Timestamp    time.Time   `json:"timestamp"`
	Protocol     string      `json:"protocol"`

	FoundBy string `json:"found_by"`
}

// GetID -
func (b *BigMapDiff) GetID() string {
	return b.ID
}

// GetIndex -
func (b *BigMapDiff) GetIndex() string {
	return "bigmapdiff"
}

// ParseElasticJSON -
func (b *BigMapDiff) ParseElasticJSON(hit gjson.Result) {
	b.ID = hit.Get("_id").String()
	b.Ptr = hit.Get("_source.ptr").Int()
	b.BinPath = hit.Get("_source.bin_path").String()
	b.Key = hit.Get("_source.key").Value()
	b.KeyHash = hit.Get("_source.key_hash").String()
	b.Value = hit.Get("_source.value").String()
	b.OperationID = hit.Get("_source.operation_id").String()
	b.Level = hit.Get("_source.level").Int()
	b.Address = hit.Get("_source.address").String()
	b.Network = hit.Get("_source.network").String()
	b.IndexedTime = hit.Get("_source.indexed_time").Int()
	b.Timestamp = hit.Get("_source.timestamp").Time().UTC()
	b.Protocol = hit.Get("_source.protocol").String()

	b.KeyStrings = parseStringsArray(hit.Get("_source.key_strings").Array())
	b.ValueStrings = parseStringsArray(hit.Get("_source.value_strings").Array())

	b.FoundBy = b.FoundByName(hit)
}

// GetScores -
func (b *BigMapDiff) GetScores(search string) []string {
	return []string{
		"key_strings^8",
		"value_strings^7",
		"key_hash",
		"address",
	}
}

// FoundByName -
func (b *BigMapDiff) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := b.GetScores("")
	return getFoundBy(keys, categories)
}
