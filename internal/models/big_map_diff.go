package models

import (
	"time"

	"github.com/tidwall/gjson"
)

// BigMapDiff -
type BigMapDiff struct {
	ID          string      `json:"-"`
	Ptr         int64       `json:"ptr,omitempty"`
	BinPath     string      `json:"bin_path"`
	Key         interface{} `json:"key"`
	KeyHash     string      `json:"key_hash"`
	KeyStrings  []string    `json:"key_strings"`
	Value       string      `json:"value"`
	OperationID string      `json:"operation_id"`
	Level       int64       `json:"level"`
	Address     string      `json:"address"`
	Network     string      `json:"network"`
	IndexedTime int64       `json:"indexed_time"`
	Timestamp   time.Time   `json:"timestamp"`
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
	b.Network = hit.Get("_source.newtork").String()
	b.IndexedTime = hit.Get("_source.indexed_time").Int()
	b.Timestamp = hit.Get("_source.timestamp").Time().UTC()

	b.parseKeyStringsArray(hit)
}

func (b *BigMapDiff) parseKeyStringsArray(hit gjson.Result) {
	result := make([]string, 0)
	for _, item := range hit.Get("_source.key_strings").Array() {
		result = append(result, item.String())
	}
	b.KeyStrings = result
}
