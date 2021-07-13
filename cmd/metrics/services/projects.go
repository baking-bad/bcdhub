package services

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/pkg/errors"
)

// ProjectsHandler -
type ProjectsHandler struct {
	*config.Context
}

// NewProjectsHandler -
func NewProjectsHandler(ctx *config.Context) *ProjectsHandler {
	return &ProjectsHandler{ctx}
}

// Handle -
func (p *ProjectsHandler) Handle(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}
	contracts := make([]*contract.Contract, len(items))
	for i := range items {
		c, ok := items[i].(*contract.Contract)
		if !ok {
			return errors.Errorf("[Projects.Handle] invalid entity type: wait *contract.Contract got %T", items[i])
		}
		contracts[i] = c
	}
	updates := make([]models.Model, 0)
	for i := range contracts {
		res, err := p.process(contracts[i], contracts[:i])
		if err != nil {
			return errors.Errorf("[Projects.Handle] compute error message: %s", err)
		}

		updates = append(updates, res...)
	}

	if len(updates) == 0 {
		return nil
	}

	logger.Info().Msgf("%2d contracts are processed", len(updates))

	if err := p.Storage.Save(updates); err != nil {
		return err
	}
	return saveSearchModels(p.Context, updates)
}

// Chunk -
func (p *ProjectsHandler) Chunk(lastID, size int64) ([]models.Model, error) {
	contracts, err := getContracts(p.StorageDB.DB, lastID, size)
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(contracts))
	for i := range contracts {
		data[i] = &contracts[i]
	}
	return data, nil
}

func (p *ProjectsHandler) process(contract *contract.Contract, chunk []*contract.Contract) ([]models.Model, error) {
	if contract.ProjectID != "" {
		return nil, nil
	}

	if err := metrics.SetContractProjectID(p.Contracts, contract, chunk); err != nil {
		return nil, errors.Errorf("error during set contract projectID: %s", err)
	}

	return []models.Model{contract}, nil
}
