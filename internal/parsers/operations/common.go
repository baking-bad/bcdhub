package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/balanceupdate"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// Metadata -
type Metadata struct {
	Result         operation.Result
	BalanceUpdates []*balanceupdate.BalanceUpdate
}

func parseMetadata(item gjson.Result, operation operation.Operation) *Metadata {
	path := "metadata.operation_result"
	if !item.Get(path).Exists() {
		path = "result"
		if !item.Get(path).Exists() {
			return nil
		}
	}

	return &Metadata{
		BalanceUpdates: NewBalanceUpdate(path, operation).Parse(item),
		Result:         NewResult(path).Parse(item),
	}
}
