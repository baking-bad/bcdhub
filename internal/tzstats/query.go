package tzstats

import (
	"fmt"
	"strings"
)

// Query -
type Query struct {
	selectedTable string
	params        map[string]string

	api *TzStats
}

// Query - request to TzStats
func (q *Query) Query(response interface{}) error {
	err := q.api.GetTable(q.selectedTable, q.params, response)
	return err
}

// Get - request to TzStats
func (q *Query) Get() (TableResponse, error) {
	var response TableResponse
	err := q.api.GetTable(q.selectedTable, q.params, &response)
	return response, err
}

// Count - request rows count
func (q *Query) Count() (int, error) {
	q.params["columns"] = "row_id"
	var response int
	err := q.api.GetTable(q.selectedTable, q.params, &response)
	return response, err
}

// Columns - Add requested columns to request
func (q *Query) Columns(columns []string) *Query {
	q.params["columns"] = strings.Join(columns, ",")
	return q
}

// Limit - Add limit items to request
func (q *Query) Limit(limit int) *Query {
	q.params["limit"] = fmt.Sprintf("%d", limit)
	return q
}

// Order - Result order either asc (default) or desc, sorted by row_id
func (q *Query) Order(order string) *Query {
	q.params["order"] = order
	return q
}

// Is - Matches rows where column values match exactly the filter value
func (q *Query) Is(column string, value string) *Query {
	q.params[column] = value
	return q
}

// NotEquals - Matches rows where column values do not match the filter value.
func (q *Query) NotEquals(column string, value string) *Query {
	column = fmt.Sprintf("%s.ne", column)
	q.params[column] = value
	return q
}

// GreaterThan - Matches columns who’s value is strictly greater than the filter value.
func (q *Query) GreaterThan(column string, value int) *Query {
	column = fmt.Sprintf("%s.gt", column)
	q.params[column] = fmt.Sprintf("%d", value)
	return q
}

// GreaterThanOrEqual  - Matches columns who’s value is greater than or equal to the filter value.
func (q *Query) GreaterThanOrEqual(column string, value int) *Query {
	column = fmt.Sprintf("%s.gte", column)
	q.params[column] = fmt.Sprintf("%d", value)
	return q
}

// LessThan - Matches columns who’s value is strictly smaller than the filter value.
func (q *Query) LessThan(column string, value int) *Query {
	column = fmt.Sprintf("%s.lt", column)
	q.params[column] = fmt.Sprintf("%d", value)
	return q
}

// LessThanOrEqual  - Matches columns who’s value is strictly smaller than or equal to the filter value.
func (q *Query) LessThanOrEqual(column string, value int) *Query {
	column = fmt.Sprintf("%s.lte", column)
	q.params[column] = fmt.Sprintf("%d", value)
	return q
}

// In  - Matches columns who’s value is equal to one of the filter values. Multiple values must be separated by comma.
func (q *Query) In(column string, values []string) *Query {
	column = fmt.Sprintf("%s.in", column)
	q.params[column] = strings.Join(values, ",")
	return q
}

// NotIn  - Matches columns who’s value is not equal to one of the filter values. Multiple values may be separated by comma.
func (q *Query) NotIn(column string, values []string) *Query {
	column = fmt.Sprintf("%s.nin", column)
	q.params[column] = strings.Join(values, ",")
	return q
}

// Range  - Matches columns who’s value is between the provided filter values, boundary inclusive. Requires exactly two values separated by comma. (This is similar to, but faster than using .gte= and .lte= in combination.)
func (q *Query) Range(column string, start, end float64) *Query {
	column = fmt.Sprintf("%s.rg", column)
	q.params[column] = fmt.Sprintf("%.f,%.f", start, end)
	return q
}

// RegExp  - Matches columns who’s value matches the regular expression. Can only be used on string-type columns (not enum or hash). Non-URL-safe characters must be properly escaped.
func (q *Query) RegExp(column string, expr string) *Query {
	column = fmt.Sprintf("%s.re", column)
	q.params[column] = expr
	return q
}
