package core

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/pkg/errors"
)

const (
	histogramRequestTemplate = `
		with f as (
			select generate_series(
			date_trunc('%s', date '2018-06-25'),
			date_trunc('%s', now()),
			'1 %s'::interval
			) as val
		)
		select
			extract(epoch from f.val),
			%s
		from f
		left join %s on date_trunc('%s', %s.timestamp) = f.val %s
		group by 1
		order by date_part
	`
)

func buildHistogramContext(ctx models.HistogramContext) (string, error) {
	var f string
	if ctx.HasFunction() {
		switch ctx.Function.Name {
		case "sum":
			f = fmt.Sprintf("sum(%s) as value", ctx.Function.Field)
		case "cardinality":
			f = fmt.Sprintf("count(distinct(%s)) as value", ctx.Function.Field)
		default:
			return "", errors.Errorf("Invalid function: %s", ctx.Function.Name)
		}
	} else {
		f = "count(id) as value"
	}

	var conditions string
	if len(ctx.Filters) > 0 {
		conds := make([]string, 0)
		for i, fltr := range ctx.Filters {
			switch fltr.Kind {
			case models.HistogramFilterKindExists:
				conds = append(conds, fmt.Sprintf("(%s is not null and %s != '')", fltr.Field, fltr.Field))
			case models.HistogramFilterKindMatch:
				if s, ok := fltr.Value.(string); ok {
					conds = append(conds, fmt.Sprintf("(%s = '%s')", fltr.Field, s))
				} else {
					conds = append(conds, fmt.Sprintf("(%s = %v)", fltr.Field, fltr.Value))
				}
			case models.HistogramFilterKindIn:
				if arr, ok := fltr.Value.([]string); ok {
					conds = append(conds, fmt.Sprintf("(array['%s'] && %s)", strings.Join(arr, "','"), fltr.Field))
				}
			case models.HistogramFilterKindAddresses:
				if value, ok := fltr.Value.([]string); ok {
					addresses := make([]string, 0)
					for j := range value {
						addresses = append(addresses, fmt.Sprintf("(%s = '%s')", fltr.Field, value[j]))
					}
					conds = append(conds, fmt.Sprintf("(%s)", strings.Join(addresses, " or ")))
				}
			case models.HistogramFilterDexEnrtypoints:
				if value, ok := fltr.Value.([]dapp.DAppContract); ok {
					entrypoints := make([]string, 0)
					for _, val := range value {
						for j := range value[i].Entrypoint {
							s := fmt.Sprintf("(intiator = '%s' and parent = '%s')", val.Address, val.Entrypoint[j])
							entrypoints = append(entrypoints, s)
						}
					}
					conds = append(conds, fmt.Sprintf("(%s)", strings.Join(entrypoints, " or ")))
				}
			}
		}

		conditions = fmt.Sprintf("where (%s)", strings.Join(conds, " and "))
	}

	return getRequest(ctx.Period, ctx.Index, f, conditions)
}

// HistogramResponse -
type HistogramResponse struct {
	DatePart float64
	Value    int64
}

// GetDateHistogram -
func (p *Postgres) GetDateHistogram(period string, opts ...models.HistogramOption) ([][]int64, error) {
	ctx := models.HistogramContext{
		Period: period,
	}
	for _, opt := range opts {
		opt(&ctx)
	}

	req, err := buildHistogramContext(ctx)
	if err != nil {
		return nil, err
	}

	// log.Print(req)

	var res []HistogramResponse
	if err := p.DB.Raw(req).Scan(&res).Error; err != nil {
		return nil, err
	}
	hist := make([][]int64, 0, len(res))
	for i := range res {
		hist = append(hist, []int64{int64(res[i].DatePart * 1000), res[i].Value})
	}

	return hist, nil
}

func getRequest(period, table, f, conditions string) (string, error) {
	if err := ValidateHistogramPeriod(period); err != nil {
		return "", err
	}

	if !helpers.StringInArray(table, []string{models.DocContracts, models.DocOperations, models.DocTransfers}) {
		return "", errors.Errorf("Invalid table: %s", table)
	}

	return fmt.Sprintf(histogramRequestTemplate, period, period, period, f, table, period, table, conditions), nil
}

// ValidateHistogramPeriod -
func ValidateHistogramPeriod(period string) error {
	if !helpers.StringInArray(period, []string{"day", "week", "month", "year"}) {
		return errors.Errorf("Invalid period: %s", period)
	}
	return nil
}
