package indexer

import "github.com/baking-bad/bcdhub/internal/models"

// Initializer -
type Initializer struct {
	repo models.GeneralRepository
}

// NewInitializer -
func NewInitializer(repo models.GeneralRepository) Initializer {
	return Initializer{repo}
}

// Init -
// TODO: create indeices
func (initializer Initializer) Init() error {
	return initializer.repo.CreateTables()
}
