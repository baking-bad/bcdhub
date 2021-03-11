package handlers

import "github.com/baking-bad/bcdhub/internal/models"

// Handler -
type Handler interface {
	Do(model models.Model) (bool, []models.Model, error)
}
