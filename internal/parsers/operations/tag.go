package operations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

func setTags(ctx *config.Context, contract *contract.Contract, op *operation.Operation) error {
	if op.Destination.Type != types.AccountTypeContract {
		return nil
	}

	var tags types.Tags
	if contract == nil {
		c, err := ctx.Cache.ContractTags(op.Destination.Address)
		if err != nil {
			if ctx.Storage.IsRecordNotFound(err) {
				return nil
			}
			return err
		}
		tags = c
	} else {
		tags = contract.Tags
	}

	if tags.Has(types.FA12Tag) {
		op.Tags.Set(types.FA12Tag)
	}
	if tags.Has(types.FA2Tag) {
		op.Tags.Set(types.FA2Tag)
	}
	if tags.Has(types.LedgerTag) {
		op.Tags.Set(types.LedgerTag)
	}
	return nil
}
