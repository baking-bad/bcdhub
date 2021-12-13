package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	year  = "year"
	month = "month"
	week  = "week"
	hour  = "hour"
	day   = "day"
	all   = "all"
)

const (
	histogramRequestTemplate = `
		with f as (
			select generate_series(
			date_trunc(?, %s),
			date_trunc(?, now()),
			?::interval
			) as val
		)
		select
			extract(epoch from f.val),
			%s
		from f
		left join %s on date_trunc(?, %s.timestamp) = f.val %s
		where  %s.id is not null
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
		f = "count(*) as value"
	}

	var conditions string
	if len(ctx.Filters) > 0 {
		conds := make([]string, 0)
		for i, fltr := range ctx.Filters {
			switch fltr.Kind {
			case models.HistogramFilterKindExists:
				conds = append(conds, fmt.Sprintf("(%s is not null and %s != '')", fltr.Field, fltr.Field))
			case models.HistogramFilterKindMatch:
				switch val := fltr.Value.(type) {
				case string:
					conds = append(conds, fmt.Sprintf("(%s = '%s')", fltr.Field, val))
				case types.Network, types.OperationStatus:
					conds = append(conds, fmt.Sprintf("(%s = %d)", fltr.Field, val))
				default:
					conds = append(conds, fmt.Sprintf("(%s = %v)", fltr.Field, val))
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

		conditions = fmt.Sprintf("and (%s)", strings.Join(conds, " and "))
	}

	return getRequest(ctx.Period, ctx.Index, f, conditions)
}

// HistogramResponse -
type HistogramResponse struct {
	DatePart float64
	Value    float64
}

func (res HistogramResponse) toFloat64() []float64 {
	return []float64{res.DatePart * 1000, res.Value}
}

// GetDateHistogram -
func (p *Postgres) GetDateHistogram(period string, opts ...models.HistogramOption) ([][]float64, error) {
	if err := ValidateHistogramPeriod(period); err != nil {
		return nil, err
	}

	ctx := models.HistogramContext{}
	for _, opt := range opts {
		opt(&ctx)
	}

	req, err := buildHistogramContext(ctx)
	if err != nil {
		return nil, err
	}

	periodName := name(period)

	var res []HistogramResponse
	if err := p.DB.Raw(req, periodName, periodName, fmt.Sprintf("1 %s", periodName), periodName).Scan(&res).Error; err != nil {
		return nil, err
	}
	hist := make([][]float64, 0, len(res))
	for i := range res {
		hist = append(hist, res[i].toFloat64())
	}

	return hist, nil
}

// GetCachedHistogram -
func (p *Postgres) GetCachedHistogram(period, name, network string) ([][]float64, error) {
	var res []HistogramResponse
	if err := p.DB.Table("series_?_by_?_?", gorm.Expr(name), gorm.Expr(period), gorm.Expr(network)).Limit(limit(period)).Order("date_part desc").Find(&res).Error; err != nil {
		return nil, err
	}
	hist := make([][]float64, 0, len(res))
	for i := range res {
		hist = append(hist, res[i].toFloat64())
	}
	return hist, nil
}

func getRequest(period, table, f, conditions string) (string, error) {
	if !helpers.StringInArray(table, []string{models.DocContracts, models.DocOperations, models.DocTransfers}) {
		return "", errors.Errorf("Invalid table: %s", table)
	}

	from := GetHistogramInterval(period)
	return fmt.Sprintf(histogramRequestTemplate, from, f, table, table, conditions, table), nil
}

// ValidateHistogramPeriod -
func ValidateHistogramPeriod(period string) error {
	if !helpers.StringInArray(period, []string{day, week, month, year, hour, all}) {
		return errors.Errorf("Invalid period: %s", period)
	}
	return nil
}

// GetHistogramInterval -
func GetHistogramInterval(period string) string {
	switch period {
	case hour:
		return "now() - interval '23 hour'" // -1 hour/day/week/month because postgres series count current date. In maths: [from; to] -> (from; to]
	case day:
		return "now() - interval '30 day'"
	case week:
		return "now() - interval '15 week'"
	case month:
		return "now() - interval '11 month'"
	default:
		return "date '2018-06-25'"
	}
}

func limit(period string) int {
	switch period {
	case hour:
		return 24
	case day:
		return 30
	case week:
		return 14
	case month:
		return 12
	case all:
		now := time.Now()
		years := now.Year() - 2018
		months := now.Month() + 6
		return years*12 + int(months)
	default:
		return 60
	}
}

func name(period string) string {
	switch period {
	case hour:
		return period
	case day:
		return period
	case week:
		return period
	case month:
		return period
	default:
		return month
	}
}
