package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

func setTags(ctx *config.Context, op *operation.Operation) error {
	if !bcd.IsContract(op.Destination) {
		return nil
	}

	c, err := ctx.CachedContract(op.Network, op.Destination)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	if c == nil {
		return nil
	}

	if c.Tags.Has(types.FA12Tag) {
		op.Tags.Set(types.FA12Tag)
	}
	if c.Tags.Has(types.FA2Tag) {
		op.Tags.Set(types.FA2Tag)
	}
	if c.Tags.Has(types.LedgerTag) {
		op.Tags.Set(types.LedgerTag)
	}
	return nil
}
