package domains

import (
	"errors"
	"html/template"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
				{{- range $index, $network := .Networks }}
				{{- if (gt $index 0) }}
				union all
				{{- end}}
				select '{{ $network }}' as network, contracts.*, accounts.address as account__address from {{ $network }}.contracts
				join {{ $network }}.accounts on contracts.account_id = accounts.id
				left join {{$network}}.scripts as alpha on alpha.id = contracts.alpha_id
				left join {{$network}}.scripts as babylon on babylon.id = contracts.babylon_id
				left join {{$network}}.scripts as jakarta on jakarta.id = contracts.jakarta_id
				where (alpha.hash = '{{$.Hash}}' or babylon.hash = '{{$.Hash}}' or jakarta.hash = '{{$.Hash}}')
				{{end}}
			) as q
			where NOT (network = '{{.network}}' and id = {{.ID}})
		) as same
		limit {{.limit}}
		offset {{.offset}}`,
	),
)

// Same -
func (storage *Storage) Same(network string, c contract.Contract, limit, offset int, availiableNetworks ...string) ([]domains.Same, error) {
	if limit < 1 || limit > 10 {
		limit = 10
	}

	if offset < 1 {
		offset = 0
	}

	if len(availiableNetworks) == 0 {
		availiableNetworks = []string{types.Mainnet.String()}
	}

	script := c.CurrentScript()
	if script == nil {
		return nil, errors.New("invalid contract script")
	}

	data := map[string]any{
		"ID":       c.ID,
		"Hash":     script.Hash,
		"limit":    limit,
		"offset":   offset,
		"network":  network,
		"Networks": availiableNetworks,
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
			{{- range $index, $network := .Networks }}
			{{- if (gt $index 0) }}
			union all
			{{- end}}
			select count(*) as c from {{$network}}.contracts
			left join {{$network}}.scripts as alpha on alpha.id = contracts.alpha_id
			left join {{$network}}.scripts as babylon on babylon.id = contracts.babylon_id
			left join {{$network}}.scripts as jakarta on jakarta.id = contracts.jakarta_id
			where (alpha.hash = '{{$.Hash}}' or babylon.hash = '{{$.Hash}}' or jakarta.hash = '{{$.Hash}}')
			{{end}}
		) as same`,
	),
)

// SameCount -
func (storage *Storage) SameCount(c contract.Contract, availiableNetworks ...string) (int, error) {
	script := c.CurrentScript()
	if script == nil {
		return 0, errors.New("invalid contract script")
	}

	data := map[string]any{
		"ID":       c.ID,
		"Hash":     script.Hash,
		"Networks": availiableNetworks,
	}

	var buffer strings.Builder
	if err := sameCountTemplate.Execute(&buffer, data); err != nil {
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
