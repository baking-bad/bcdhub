package transfer

import (
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/restream/reindexer"
)

func buildGetContext(ctx transfer.GetContext, query *reindexer.Query) {
	buildGetContextWithoutLimits(ctx, query)

	appendSort(ctx, query)
	appendOffset(ctx, query)
	appendLimit(ctx, query)
}

func buildGetContextWithoutLimits(ctx transfer.GetContext, query *reindexer.Query) {
	filterNetwork(ctx, query)
	filterAddress(ctx, query)
	filterTime(ctx, query)
	filterCursor(ctx, query)
	filterContracts(ctx, query)
	filterTokenID(ctx, query)
	filterHash(ctx, query)
}

func filterNetwork(ctx transfer.GetContext, query *reindexer.Query) {
	if ctx.Network != "" {
		query = query.Match("network", ctx.Network)
	}
}

func filterHash(ctx transfer.GetContext, query *reindexer.Query) {
	if ctx.Hash != "" {
		query = query.Match("hash", ctx.Hash)
	}
}

func filterAddress(ctx transfer.GetContext, query *reindexer.Query) {
	if ctx.Address == "" {
		return
	}
	query = query.OpenBracket().
		Match("from", ctx.Address).
		Or().
		Match("to", ctx.Address).
		CloseBracket()
}

func filterTokenID(ctx transfer.GetContext, query *reindexer.Query) {
	if ctx.TokenID >= 0 {
		query = query.WhereInt64("token_id", reindexer.EQ, ctx.TokenID)
	}
}

func filterTime(ctx transfer.GetContext, query *reindexer.Query) {
	if ctx.Start > 0 {
		query = query.WhereInt64("timestamp", reindexer.GE, int64(ctx.Start))
	}
	if ctx.End > 0 {
		query = query.WhereInt64("timestamp", reindexer.LT, int64(ctx.End))
	}
}

func filterCursor(ctx transfer.GetContext, query *reindexer.Query) {
	if ctx.LastID != "" {
		condition := reindexer.LT
		if ctx.SortOrder == "asc" {
			condition = reindexer.GT
		}
		query = query.Where("indexed_time", condition, ctx.LastID)
	}
}

func filterContracts(ctx transfer.GetContext, query *reindexer.Query) {
	if len(ctx.Contracts) == 0 {
		return
	}

	query = query.OpenBracket()

	for i := range ctx.Contracts {
		query = query.Match("contract", ctx.Contracts[i])
		if len(ctx.Contracts)-1 > i {
			query = query.Or()
		}
	}

	query = query.CloseBracket()
}

func appendLimit(ctx transfer.GetContext, query *reindexer.Query) {
	if ctx.Size > 0 && ctx.Size <= maxTransfersSize {
		query = query.Limit(int(ctx.Size))
	} else {
		query = query.Limit(maxTransfersSize)
	}
}

func appendOffset(ctx transfer.GetContext, query *reindexer.Query) {
	if ctx.Offset > 0 && ctx.Offset <= maxTransfersSize {
		query = query.Offset(int(ctx.Offset))
	}
}

func appendSort(ctx transfer.GetContext, query *reindexer.Query) {
	query = query.Sort("timestamp", ctx.SortOrder == "desc")
}
