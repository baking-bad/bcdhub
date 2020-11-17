package handlers

import "github.com/baking-bad/bcdhub/internal/elastic"

// Handler -
type Handler interface {
	Do(model elastic.Model) (bool, error)
}
