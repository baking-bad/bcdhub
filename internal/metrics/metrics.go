package metrics

import (
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
)

// Handler -
type Handler struct {
	ES elastic.IElastic
	DB database.DB
}

// New -
func New(es elastic.IElastic, db database.DB) *Handler {
	return &Handler{es, db}
}
