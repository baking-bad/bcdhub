package metrics

import (
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
)

// Handler -
type Handler struct {
	ES *elastic.Elastic
	DB database.DB
}

// New -
func New(es *elastic.Elastic, db database.DB) *Handler {
	return &Handler{
		ES: es,
		DB: db,
	}
}
