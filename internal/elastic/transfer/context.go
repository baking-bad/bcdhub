package transfer

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/helpers"
)

// GetTransfersContext -
type GetTransfersContext struct {
	Contracts []string
	Network   string
	Address   string
	Hash      string
	Start     uint
	End       uint
	SortOrder string
	LastID    string
	Size      int64
	Offset    int64
	TokenID   int64

	query   core.Base
	filters []core.Item
}

// Build -
func (ctx *GetTransfersContext) Build() interface{} {
	ctx.query = core.NewQuery()
	ctx.filters = make([]core.Item, 0)

	ctx.filterNetwork()
	ctx.filterAddress()
	ctx.filterTime()
	ctx.filterCursor()
	ctx.filterContracts()
	ctx.filterTokenID()
	ctx.filterHash()

	ctx.query.Query(
		core.Bool(
			core.Filter(
				ctx.filters...,
			),
		),
	)
	ctx.appendSort()
	ctx.appendOffset()
	ctx.appendSize()
	return ctx.query
}

func (ctx *GetTransfersContext) filterNetwork() {
	if ctx.Network != "" {
		ctx.filters = append(ctx.filters, core.Match("network", ctx.Network))
	}
}

func (ctx *GetTransfersContext) filterHash() {
	if ctx.Hash != "" {
		ctx.filters = append(ctx.filters, core.MatchPhrase("hash", ctx.Hash))
	}
}

func (ctx *GetTransfersContext) filterAddress() {
	if ctx.Address == "" {
		return
	}

	ctx.filters = append(ctx.filters, core.Bool(
		core.Should(
			core.MatchPhrase("from", ctx.Address),
			core.MatchPhrase("to", ctx.Address),
		),
		core.MinimumShouldMatch(1),
	))
}

func (ctx *GetTransfersContext) filterTokenID() {
	if ctx.TokenID >= 0 {
		ctx.filters = append(ctx.filters, core.Term("token_id", ctx.TokenID))
	}
}

func (ctx *GetTransfersContext) filterTime() {
	ts := core.Item{}
	if ctx.Start > 0 {
		ts["gte"] = ctx.Start
	}
	if ctx.End > 0 {
		ts["lt"] = ctx.End
	}
	if len(ts) > 0 {
		ctx.filters = append(ctx.filters, core.Range("timestamp", ts))
	}
}

func (ctx *GetTransfersContext) filterCursor() {
	if ctx.LastID != "" {
		eq := "lt"
		if ctx.SortOrder == "asc" {
			eq = "gt"
		}
		ctx.filters = append(ctx.filters, core.Range("indexed_time", core.Item{eq: ctx.LastID}))
	}
}

func (ctx *GetTransfersContext) filterContracts() {
	if len(ctx.Contracts) == 0 {
		return
	}

	shouldItems := make([]core.Item, len(ctx.Contracts))
	for i := range ctx.Contracts {
		shouldItems[i] = core.MatchPhrase("contract", ctx.Contracts[i])
	}

	ctx.filters = append(ctx.filters, core.Bool(
		core.Should(shouldItems...),
		core.MinimumShouldMatch(1),
	))
}

func (ctx *GetTransfersContext) appendSize() {
	if ctx.Size > 0 && ctx.Size <= maxTransfersSize {
		ctx.query.Size(ctx.Size)
	} else {
		ctx.query.Size(maxTransfersSize)
	}
}

func (ctx *GetTransfersContext) appendOffset() {
	if ctx.Offset > 0 && ctx.Offset <= maxTransfersSize {
		ctx.query.From(ctx.Offset)
	}
}

func (ctx *GetTransfersContext) appendSort() {
	if helpers.StringInArray(ctx.SortOrder, []string{"desc", "asc"}) {
		ctx.query.Sort("timestamp", ctx.SortOrder)
	} else {
		ctx.query.Sort("timestamp", "desc")
	}
}
