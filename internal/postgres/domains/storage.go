package domains

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"gorm.io/gorm/clause"
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
	select tb.network, tb.contract, tb.token_id, tb.balance, tm.symbol, tm.name, tm.decimals, tm.description, tm.artifact_uri, tm.display_uri, tm.external_uri, tm.thumbnail_uri, tm.is_transferable, tm.is_boolean_amount, tm.should_prefer_symbol, tm.tags, tm.creators, tm.formats, tm.extras  from (
		(?)  as tb
		left join token_metadata as tm on tm.network  = tb.network and tm.token_id = tb.token_id and tm.contract = tb.contract
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
func (storage *Storage) TokenBalances(network types.Network, contract, address string, size, offset int64, sort string, hideZeroBalances bool) (domains.TokenBalanceResponse, error) {
	response := domains.TokenBalanceResponse{
		Balances: make([]domains.TokenBalance, 0),
	}

	query := storage.DB.Table(models.DocTokenBalances).Scopes(core.NetworkAndAddress(network, address))

	if contract != "" {
		query.Where("contract = ?", contract)
	}

	if hideZeroBalances {
		query.Where("balance != 0")
	}

	if err := query.Count(&response.Count).Error; err != nil {
		return response, err
	}

	query.Limit(storage.getPageSizeForBalances(size)).Offset(int(offset))

	switch sort {
	case "token_id":
	case "balance":
		query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: "balance"},
			Desc:   true,
		})
		query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: "id"},
			Desc:   true,
		})
	default:
		query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: "token_id"},
			Desc:   true,
		})
	}

	if err := storage.DB.Raw(balanceQuery, query).
		Find(&response.Balances).Error; err != nil {
		return response, err
	}

	return response, nil
}

var transfersQuery = `
	select o.hash, o.nonce, o.counter, t.*, tm.symbol, tm.decimals, tm."name"  from (
		(?)  as t
		left join operations as o on o.id  = t.operation_id
		left join token_metadata tm on tm.network = t.network and tm.contract = t.contract and tm.token_id = t.token_id
	)
`

// Transfers -
func (storage *Storage) Transfers(ctx transfer.GetContext) (domains.TransfersResponse, error) {
	response := domains.TransfersResponse{
		Transfers: make([]domains.Transfer, 0),
	}
	query := storage.DB.Table(models.DocTransfers)
	storage.buildGetContext(query, ctx, true)

	if err := storage.DB.Raw(transfersQuery, query).Find(&response.Transfers).Error; err != nil {
		return response, err
	}

	received := len(response.Transfers)
	size := storage.GetPageSize(ctx.Size)
	if ctx.Offset == 0 && size > received {
		response.Total = int64(len(response.Transfers))
	} else {
		countQuery := storage.DB.Table(models.DocTransfers)
		storage.buildGetContext(countQuery, ctx, false)
		if err := countQuery.Count(&response.Total).Error; err != nil {
			return response, err
		}
	}

	if received > 0 {
		response.LastID = fmt.Sprintf("%d", response.Transfers[received-1].ID)
	}
	return response, nil
}

// BigMapDiffs -
func (storage *Storage) BigMapDiffs(lastID, size int64) (result []domains.BigMapDiff, err error) {
	query := storage.DB.Table(models.DocBigMapDiff).Preload("Operation").Preload("Protocol").Order("id asc")
	if lastID > 0 {
		query.Where("id > ?", lastID)
	}
	query.Limit(storage.GetPageSize(size))
	err = query.Find(&result).Error
	return
}
