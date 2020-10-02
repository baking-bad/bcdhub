package elastic

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
)

// GetTransfersContext -
type GetTransfersContext struct {
	Contracts []string
	Network   string
	From      string
	To        string
	SortOrder string
	LastID    string
	Size      int64
	Offset    int64
	TokenID   int64

	query   base
	filters []qItem
}

func (ctx GetTransfersContext) buildQuery() base {
	ctx.query = newQuery()
	ctx.filters = make([]qItem, 0)

	ctx.filterNetwork()
	ctx.filterAddresses()
	ctx.filterCursor()
	ctx.filterContracts()
	ctx.filterTokenID()

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

func (ctx GetTransfersContext) filterNetwork() {
	if ctx.Network != "" {
		ctx.filters = append(ctx.filters, matchQ("network", ctx.Network))
	}
}

func (ctx GetTransfersContext) filterTokenID() {
	if ctx.TokenID >= 0 {
		ctx.filters = append(ctx.filters, term("token_id", ctx.TokenID))
	}
}

func (ctx GetTransfersContext) filterAddresses() {
	if ctx.From != "" || ctx.To != "" {
		shouldItems := make([]qItem, 0)
		if ctx.From != "" {
			shouldItems = append(shouldItems, matchPhrase("from", ctx.From))
		}
		if ctx.To != "" {
			shouldItems = append(shouldItems, matchPhrase("to", ctx.To))
		}
		ctx.filters = append(ctx.filters, boolQ(
			should(shouldItems...),
			minimumShouldMatch(1),
		))
	}
}

func (ctx GetTransfersContext) filterCursor() {
	if ctx.LastID != "" {
		eq := "lt"
		if ctx.SortOrder == "asc" {
			eq = "gt"
		}
		ctx.filters = append(ctx.filters, rangeQ("indexed_time", qItem{eq: ctx.LastID}))
	}
}

func (ctx GetTransfersContext) filterContracts() {
	if len(ctx.Contracts) == 0 {
		return
	}

	ctx.filters = append(ctx.filters, in("contract", ctx.Contracts))
}

func (ctx GetTransfersContext) appendSize() {
	if ctx.Size > 0 && ctx.Size <= 100 {
		ctx.query.Size(ctx.Size)
	} else {
		ctx.query.Size(defaultSize)
	}
}

func (ctx GetTransfersContext) appendOffset() {
	if ctx.Offset > 0 && ctx.Offset <= 100 {
		ctx.query.Offset(ctx.Offset)
	}
}

func (ctx GetTransfersContext) appendSort() {
	if helpers.StringInArray(ctx.SortOrder, []string{"desc", "asc"}) {
		ctx.query.Sort("indexed_time", ctx.SortOrder)
	} else {
		ctx.query.Sort("indexed_time", "desc")
	}
}

// GetTransfers -
func (e *Elastic) GetTransfers(ctx GetTransfersContext) (TransfersResponse, error) {
	query := ctx.buildQuery()

	po := TransfersResponse{}
	result, err := e.query([]string{DocTransfers}, query)
	if err != nil {
		return po, err
	}

	hits := result.Get("hits.hits").Array()
	transfers := make([]models.Transfer, len(hits))
	for i, hit := range hits {
		transfers[i].ParseElasticJSON(hit)
		if i == len(hits)-1 {
			po.LastID = transfers[i].ID
		}
	}
	po.Transfers = transfers
	po.Total = result.Get("hits.total.value").Int()
	return po, nil
}
