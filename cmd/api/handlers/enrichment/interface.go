package enrichment

import "github.com/tidwall/gjson"

// Enrichment -
type Enrichment interface {
	Do(string, gjson.Result) (gjson.Result, error)
	Level() int64
}
