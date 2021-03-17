package search

// Searcher -
type Searcher interface {
	ByText(string, int64, []string, map[string]interface{}, bool) (Result, error)
}
