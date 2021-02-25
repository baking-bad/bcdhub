package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// Metadata -
type Metadata struct {
	Result operation.Result
}

func parseMetadata(item gjson.Result) *Metadata {
	path := "metadata.operation_result"
	if !item.Get(path).Exists() {
		path = "result"
		if !item.Get(path).Exists() {
			return nil
		}
	}

	return &Metadata{
		Result: NewResult(path).Parse(item),
	}
}
