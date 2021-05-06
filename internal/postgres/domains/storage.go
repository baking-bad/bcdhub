package domains

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/domains"
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
	select tb.network, tb.contract, tb.token_id, tb.balance, tm.symbol, tm.name, tm.decimals, tm.description, tm.artifact_uri, tm.display_uri, tm.external_uri, tm.thumbnail_uri, tm.is_transferable, tm.is_boolean_amount, tm.should_prefer_symbol, tm.tags, tm.creators, tm.formats, tm.extras  from (
		(?)  as tb
		left join token_metadata as tm on tm.network  = tb.network and tm.token_id = tb.token_id and tm.contract = tb.contract
	)
	order by %s desc
	limit ?
	offset ?
`

// TokenBalances -
func (storage *Storage) TokenBalances(network, contract, address string, size, offset int64, sort string) (domains.TokenBalanceResponse, error) {
	response := domains.TokenBalanceResponse{
		Balances: make([]domains.TokenBalance, 0),
	}

	query := storage.DB.Table(models.DocTokenBalances).Scopes(core.NetworkAndAddress(network, address))

	if contract != "" {
		query.Where("contract = ?", contract)
	}

	if sort != "token_id" && sort != "balance" {
		sort = "token_id"
	}

	limit := storage.GetPageSize(size)
	if err := storage.DB.Raw(fmt.Sprintf(balanceQuery, sort), query, limit, offset).
		Find(&response.Balances).Error; err != nil {
		return response, err
	}

	if err := query.Count(&response.Count).Error; err != nil {
		return response, err
	}

	return response, nil
}
