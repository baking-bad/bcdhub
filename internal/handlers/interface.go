package handlers

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
)

// Handler -
type Handler interface {
	Do(bmd *bigmapdiff.BigMapDiff) (bool, []models.Model, error)
}
