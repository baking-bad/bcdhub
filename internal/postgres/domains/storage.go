package domains

import (
	"errors"
	"html/template"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
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

var sameTemplate = template.Must(
	template.New("same").Parse(
		`select * from (
			select * from (
				select 'mainnet' as network, c1.*, accounts.address as account__address from mainnet.contracts as c1
					join mainnet.accounts on c1.account_id = accounts.id
					where (c1.alpha_id = {{.AlphaID}} or c1.babylon_id = {{.BabylonID}} or c1.jakarta_id = {{.JakartaID}})
				union all
				select 'ghostnet' as network, c2.*, accounts.address as account__address from ghostnet.contracts as c2
					join ghostnet.accounts on c2.account_id = accounts.id
					where (c2.alpha_id = {{.AlphaID}} or c2.babylon_id = {{.BabylonID}} or c2.jakarta_id = {{.JakartaID}})
				union all
				select  'jakartanet' as network, c3.*, accounts.address as account__address from jakartanet.contracts as c3
					join jakartanet.accounts on c3.account_id = accounts.id
					where (c3.alpha_id = {{.AlphaID}} or c3.babylon_id = {{.BabylonID}} or c3.jakarta_id = {{.JakartaID}})
				union all
				select  'kathmandunet' as network, c3.*, accounts.address as account__address from kathmandunet.contracts as c3
					join kathmandunet.accounts on c3.account_id = accounts.id
					where (c3.alpha_id = {{.AlphaID}} or c3.babylon_id = {{.BabylonID}} or c3.jakarta_id = {{.JakartaID}})
			) as q
			where NOT (network = '{{.network}}' and id = {{.ID}})
		) as same
		limit {{.limit}}
		offset {{.offset}}`,
	),
)

// Same -
func (storage *Storage) Same(network string, c contract.Contract, limit, offset int) ([]domains.Same, error) {
	if limit < 1 || limit > 10 {
		limit = 10
	}

	if offset < 1 {
		offset = 0
	}

	data := map[string]any{
		"ID":        c.ID,
		"AlphaID":   c.AlphaID,
		"BabylonID": c.BabylonID,
		"JakartaID": c.JakartaID,
		"limit":     limit,
		"offset":    offset,
		"network":   network,
	}

	var buffer strings.Builder
	if err := sameTemplate.Execute(&buffer, data); err != nil {
		return nil, err
	}

	var same []domains.Same
	_, err := storage.DB.Query(&same, buffer.String())
	return same, err
}

var sameCountTemplate = template.Must(
	template.New("sameCount").Parse(
		`select sum(c) from (
			select count(*) as c from mainnet.contracts as c1
				where (c1.alpha_id = {{.AlphaID}} or c1.babylon_id = {{.BabylonID}} or c1.jakarta_id = {{.JakartaID}})
			union all
			select count(*) as c from ghostnet.contracts as c3
				where (c3.alpha_id = {{.AlphaID}} or c3.babylon_id = {{.BabylonID}} or c3.jakarta_id = {{.JakartaID}})
			union all
			select count(*) as c from jakartanet.contracts as c2
				where (c2.alpha_id = {{.AlphaID}} or c2.babylon_id = {{.BabylonID}} or c2.jakarta_id = {{.JakartaID}})
		) as same`,
	),
)

// SameCount -
func (storage *Storage) SameCount(c contract.Contract) (int, error) {
	var buffer strings.Builder
	if err := sameCountTemplate.Execute(&buffer, c); err != nil {
		return 0, err
	}

	var count int
	if _, err := storage.DB.QueryOne(pg.Scan(&count), buffer.String()); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return count - 1, nil
}
