package operations

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// Metadata -
type Metadata struct {
	Result         models.OperationResult
	BalanceUpdates []*models.BalanceUpdate
}

func parseMetadata(item gjson.Result, operation models.Operation) *Metadata {
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
