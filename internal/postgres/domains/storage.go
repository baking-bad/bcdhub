package domains

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

var balanceQuery = `
	select tb.contract, tb.token_id, tb.balance, tm.symbol, tm.name, tm.decimals, tm.description, tm.artifact_uri, tm.display_uri, tm.external_uri, tm.thumbnail_uri, tm.is_transferable, tm.is_boolean_amount, tm.should_prefer_symbol, tm.tags, tm.creators, tm.formats, tm.extras  from (
		(?) as tb
		left join token_metadata as tm on tm.token_id = tb.token_id and tm.contract = tb.contract
	)
`

func (storage *Storage) getPageSizeForBalances(size int64) int {
	switch {
	case size > 50:
		return 50
	case size == 0:
		return int(storage.PageSize)
	default:
		return int(size)
	}
}

// TokenBalances -
func (storage *Storage) TokenBalances(contract string, accountID int64, size, offset int64, sort string, hideZeroBalances bool) (domains.TokenBalanceResponse, error) {
	response := domains.TokenBalanceResponse{
		Balances: make([]domains.TokenBalance, 0),
	}

	query := storage.DB.Model().Table(models.DocTokenBalances).Where("account_id = ?", accountID)

	if contract != "" {
		query.Where("contract = ?", contract)
	}

	if hideZeroBalances {
		query.Where("balance != 0")
	}

	if count, err := query.Count(); err != nil {
		return response, err
	} else {
		response.Count = int64(count)
	}

	query.Limit(storage.getPageSizeForBalances(size)).Offset(int(offset))

	switch sort {
	case "token_id":
	case "balance":
		query.Order("balance desc, id desc")
	default:
		query.Order("token_id desc")
	}

	if _, err := storage.DB.Query(&response.Balances, balanceQuery, query); err != nil {
		return response, err
	}

	return response, nil
}

var transfersQuery = `
	select o.hash, o.nonce, o.counter, t.*, tm.symbol, tm.decimals, tm."name",
		"from".id as from__id, "from".address as from__address, "from".alias as from__alias, "from".type as from__type,
		"to".id as to__id, "to".address as to__address, "to".alias as to__alias, "to".type as to__type,
		"initiator".id as initiator__id, "initiator".address as initiator__address, "initiator".alias as initiator__alias, "initiator".type as initiator__type
	from (
		(?) as t
		left join operations as o on o.id  = t.operation_id
		left join token_metadata tm on tm.contract = t.contract and tm.token_id = t.token_id
		left join "accounts" AS "from" ON "from"."id" = t."from_id" 
		left join "accounts" AS "to" ON "to"."id" = t."to_id" 
		left join "accounts" AS "initiator" ON "initiator"."id" = t."initiator_id" 	
	)
`

// Transfers -
func (storage *Storage) Transfers(ctx transfer.GetContext) (domains.TransfersResponse, error) {
	response := domains.TransfersResponse{
		Transfers: make([]domains.Transfer, 0),
	}
	query := storage.DB.Model((*transfer.Transfer)(nil))
	storage.buildGetContext(query, ctx, true)

	if _, err := storage.DB.Query(&response.Transfers, transfersQuery, query); err != nil {
		return response, err
	}

	received := len(response.Transfers)
	size := storage.GetPageSize(ctx.Size)
	if ctx.Offset == 0 && size > received {
		response.Total = int64(len(response.Transfers))
	} else {
		countQuery := storage.DB.Model((*transfer.Transfer)(nil))
		storage.buildGetContext(countQuery, ctx, false)
		if count, err := countQuery.Count(); err != nil {
			return response, err
		} else {
			response.Total = int64(count)
		}
	}

	if received > 0 {
		response.LastID = fmt.Sprintf("%d", response.Transfers[received-1].ID)
	}
	return response, nil
}

// BigMapDiffs -
func (storage *Storage) BigMapDiffs(lastID, size int64) (result []domains.BigMapDiff, err error) {
	var ids []int64
	query := storage.DB.Model((*bigmapdiff.BigMapDiff)(nil)).Column("id").Order("id asc")
	if lastID > 0 {
		query.Where("big_map_diff.id > ?", lastID)
	}
	if err = query.Limit(storage.GetPageSize(size)).Select(&ids); err != nil {
		return
	}

	if len(ids) == 0 {
		return
	}

	err = storage.DB.Model((*domains.BigMapDiff)(nil)).WhereIn("big_map_diff.id IN (?)", ids).
		Relation("Operation").Relation("Protocol").
		Select(&result)
	return
}
