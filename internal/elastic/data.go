package elastic

import (
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
)

// SearchResult -
type SearchResult struct {
	Count     int64             `json:"count"`
	Time      int64             `json:"time"`
	Contracts []models.Contract `json:"contracts"`
}

// ContractStats -
type ContractStats struct {
	TxCount     int64
	SumTxAmount int64
	LastAction  time.Time
}
