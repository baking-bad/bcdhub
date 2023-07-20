package helpers

import (
	"errors"
	"time"
)

// QuarterOf -
func QuarterOf(month time.Month) int {
	return (int(month) + 2) / 3
}

// QuarterBoundaries -
func QuarterBoundaries(current time.Time) (time.Time, time.Time, error) {
	year := current.Year()
	quarter := QuarterOf(current.Month())

	switch quarter {
	case 1:
		start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0)
		return start, end, nil
	case 2:
		start := time.Date(year, time.April, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0)
		return start, end, nil
	case 3:
		start := time.Date(year, time.July, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0)
		return start, end, nil
	case 4:
		start := time.Date(year, time.October, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0)
		return start, end, nil
	}

	return time.Now(), time.Now(), errors.New("invalid quarter")
}
