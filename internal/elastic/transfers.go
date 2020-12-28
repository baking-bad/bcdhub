package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
)

const (
	maxTransfersSize = 10000
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
	Nonce     *int64
	Counter   *int64

	query   base
	filters []qItem
}

func (ctx *GetTransfersContext) buildQuery() base {
	ctx.query = newQuery()
	ctx.filters = make([]qItem, 0)

	ctx.filterNetwork()
	ctx.filterAddress()
	ctx.filterTime()
	ctx.filterCursor()
	ctx.filterContracts()
	ctx.filterTokenID()
	ctx.filterCounter()
	ctx.filterNonce()
	ctx.filterHash()

	ctx.query.Query(
		boolQ(
			filter(
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
		ctx.filters = append(ctx.filters, matchQ("network", ctx.Network))
	}
}

func (ctx *GetTransfersContext) filterHash() {
	if ctx.Hash != "" {
		ctx.filters = append(ctx.filters, matchPhrase("hash", ctx.Hash))
	}
}

func (ctx *GetTransfersContext) filterAddress() {
	if ctx.Address == "" {
		return
	}

	ctx.filters = append(ctx.filters, boolQ(
		should(
			matchPhrase("from", ctx.Address),
			matchPhrase("to", ctx.Address),
		),
		minimumShouldMatch(1),
	))
}

func (ctx *GetTransfersContext) filterTokenID() {
	if ctx.TokenID >= 0 {
		ctx.filters = append(ctx.filters, term("token_id", ctx.TokenID))
	}
}

func (ctx *GetTransfersContext) filterCounter() {
	if ctx.Counter != nil {
		ctx.filters = append(ctx.filters, term("counter", *ctx.Counter))
	}
}

func (ctx *GetTransfersContext) filterNonce() {
	if ctx.Nonce != nil {
		ctx.filters = append(ctx.filters, term("nonce", *ctx.Nonce))
	}
}

func (ctx *GetTransfersContext) filterTime() {
	ts := qItem{}
	if ctx.Start > 0 {
		ts["gte"] = ctx.Start
	}
	if ctx.End > 0 {
		ts["lt"] = ctx.End
	}
	if len(ts) > 0 {
		ctx.filters = append(ctx.filters, rangeQ("timestamp", ts))
	}
}

func (ctx *GetTransfersContext) filterCursor() {
	if ctx.LastID != "" {
		eq := "lt"
		if ctx.SortOrder == "asc" {
			eq = "gt"
		}
		ctx.filters = append(ctx.filters, rangeQ("indexed_time", qItem{eq: ctx.LastID}))
	}
}

func (ctx *GetTransfersContext) filterContracts() {
	if len(ctx.Contracts) == 0 {
		return
	}

	shouldItems := make([]qItem, len(ctx.Contracts))
	for i := range ctx.Contracts {
		shouldItems[i] = matchPhrase("contract", ctx.Contracts[i])
	}

	ctx.filters = append(ctx.filters, boolQ(
		should(shouldItems...),
		minimumShouldMatch(1),
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

// GetTransfers -
func (e *Elastic) GetTransfers(ctx GetTransfersContext) (TransfersResponse, error) {
	query := ctx.buildQuery()

	po := TransfersResponse{}

	var response SearchResponse
	if err := e.query([]string{DocTransfers}, query, &response); err != nil {
		return po, err
	}

	hits := response.Hits.Hits
	transfers := make([]models.Transfer, len(hits))
	for i := range hits {
		if err := json.Unmarshal(hits[i].Source, &transfers[i]); err != nil {
			return po, err
		}
		transfers[i].ID = hits[i].ID
	}
	po.Transfers = transfers
	po.Total = response.Hits.Total.Value
	if len(transfers) > 0 {
		po.LastID = fmt.Sprintf("%d", transfers[len(transfers)-1].IndexedTime)
	}
	return po, nil
}

// GetAllTransfers -
func (e *Elastic) GetAllTransfers(network string, level int64) ([]models.Transfer, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				rangeQ("level", qItem{"gt": level}),
			),
		),
	)

	transfers := make([]models.Transfer, 0)
	err := e.getAllByQuery(query, &transfers)
	return transfers, err
}
