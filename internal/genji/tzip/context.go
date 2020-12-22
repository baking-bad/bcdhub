package tzip

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

func buildGetTokenMetadataContext(ctx tzip.GetTokenMetadataContext) *core.Builder {
	filters := make([]fmt.Stringer, 0)

	if ctx.Contract != "" {
		filters = append(filters, core.NewEq("address", ctx.Contract))
	}
	if ctx.Network != "" {
		filters = append(filters, core.NewEq("network", ctx.Network))
	}
	if ctx.Level.IsFilled() {
		filters = append(filters, core.BuildComparator("level", ctx.Level))
	}
	if ctx.TokenID != -1 {
		filters = append(filters, core.NewEq("tokens.static.token_id", ctx.TokenID))
	}
	return core.NewBuilder().SelectAll(models.DocTZIP).And(filters...)
}
