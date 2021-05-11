package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/operation"
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
	for _, tag := range c.Tags {
		switch tag {
		case consts.FA12Tag, consts.FA2Tag:
			if op.Tags == nil {
				op.Tags = make([]string, 0)
			}
			op.Tags = append(op.Tags, tag)
		case consts.LedgerTag:
			if op.Tags == nil {
				op.Tags = make([]string, 0)
			}
			op.Tags = append(op.Tags, tag)
		}
	}
	return nil
}
