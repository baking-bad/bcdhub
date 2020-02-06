package elastic

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// SearchResult -
type SearchResult struct {
	Count     int64             `json:"count"`
	Time      int64             `json:"time"`
	Contracts []models.Contract `json:"contracts"`
}
