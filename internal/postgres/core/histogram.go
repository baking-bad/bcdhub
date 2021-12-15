package core

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/pkg/errors"
)

const (
	year  = "year"
	month = "month"
	week  = "week"
	hour  = "hour"
	day   = "day"
	all   = "all"
)

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

// HistogramResponse -
type HistogramResponse struct {
	DatePart float64
	Value    float64
}
