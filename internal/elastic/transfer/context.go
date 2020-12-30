package transfer

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

func buildGetContext(ctx transfer.GetContext) core.Base {
	query := core.NewQuery()
	filters := make([]core.Item, 0)

	if f := filterNetwork(ctx); f != nil {
		filters = append(filters, f)
	}
	if f := filterAddress(ctx); f != nil {
		filters = append(filters, f)
	}
	if f := filterTime(ctx); f != nil {
		filters = append(filters, f)
	}
	if f := filterCursor(ctx); f != nil {
		filters = append(filters, f)
	}
	if f := filterContracts(ctx); f != nil {
		filters = append(filters, f)
	}
	if f := filterTokenID(ctx); f != nil {
		filters = append(filters, f)
	}
	if f := filterHash(ctx); f != nil {
		filters = append(filters, f)
	}
	if f := filterCounter(ctx); f != nil {
		filters = append(filters, f)
	}
	if f := filterNonce(ctx); f != nil {
		filters = append(filters, f)
	}

	query.Query(
		core.Bool(
			core.Filter(
				filters...,
			),
		),
	)
	appendSort(ctx, query)
	appendOffset(ctx, query)
	appendSize(ctx, query)
	return query
}

func filterNetwork(ctx transfer.GetContext) core.Item {
	if ctx.Network != "" {
		return core.Match("network", ctx.Network)
	}
	return nil
}

func filterHash(ctx transfer.GetContext) core.Item {
	if ctx.Hash != "" {
		return core.MatchPhrase("hash", ctx.Hash)
	}
	return nil
}

func filterAddress(ctx transfer.GetContext) core.Item {
	if ctx.Address == "" {
		return nil
	}

	return core.Bool(
		core.Should(
			core.MatchPhrase("from", ctx.Address),
			core.MatchPhrase("to", ctx.Address),
		),
		core.MinimumShouldMatch(1),
	)
}

func filterTokenID(ctx transfer.GetContext) core.Item {
	if ctx.TokenID >= 0 {
		return core.Term("token_id", ctx.TokenID)
	}
	return nil
}

func filterTime(ctx transfer.GetContext) core.Item {
	ts := core.Item{}
	if ctx.Start > 0 {
		ts["gte"] = ctx.Start
	}
	if ctx.End > 0 {
		ts["lt"] = ctx.End
	}
	if len(ts) > 0 {
		return core.Range("timestamp", ts)
	}
	return nil
}

func filterCursor(ctx transfer.GetContext) core.Item {
	if ctx.LastID != "" {
		eq := "lt"
		if ctx.SortOrder == "asc" {
			eq = "gt"
		}
		return core.Range("indexed_time", core.Item{eq: ctx.LastID})
	}
	return nil
}

func filterContracts(ctx transfer.GetContext) core.Item {
	if len(ctx.Contracts) == 0 {
		return nil
	}

	shouldItems := make([]core.Item, len(ctx.Contracts))
	for i := range ctx.Contracts {
		shouldItems[i] = core.MatchPhrase("contract", ctx.Contracts[i])
	}

	return core.Bool(
		core.Should(shouldItems...),
		core.MinimumShouldMatch(1),
	)
}

func filterCounter(ctx transfer.GetContext) core.Item {
	if ctx.Counter != nil {
		return core.Term("counter", *ctx.Counter)
	}
	return nil
}

func filterNonce(ctx transfer.GetContext) core.Item {
	if ctx.Nonce != nil {
		return core.Term("nonce", *ctx.Nonce)
	}
	return nil
}

func appendSize(ctx transfer.GetContext, query core.Base) {
	if ctx.Size > 0 && ctx.Size <= maxTransfersSize {
		query.Size(ctx.Size)
	} else {
		query.Size(maxTransfersSize)
	}
}

func appendOffset(ctx transfer.GetContext, query core.Base) {
	if ctx.Offset > 0 && ctx.Offset <= maxTransfersSize {
		query.From(ctx.Offset)
	}
}

func appendSort(ctx transfer.GetContext, query core.Base) {
	if helpers.StringInArray(ctx.SortOrder, []string{"desc", "asc"}) {
		query.Sort("timestamp", ctx.SortOrder)
	} else {
		query.Sort("timestamp", "desc")
	}
}
