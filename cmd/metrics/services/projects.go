package services

import (
	"context"

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
func (p *ProjectsHandler) Handle(ctx context.Context, items []models.Model) error {
	if len(items) == 0 {
		return nil
	}
	scripts := make([]contract.Script, len(items))
	for i := range items {
		c, ok := items[i].(*contract.Script)
		if !ok {
			return errors.Errorf("[Projects.Handle] invalid entity type: wait *contract.Script got %T", items[i])
		}
		scripts[i] = *c
	}

	for i := range scripts {
		if err := p.process(&scripts[i], scripts[:i]); err != nil {
			return errors.Errorf("[Projects.Handle] compute error message: %s", err)
		}
	}

	if len(scripts) > 0 {
		logger.Info().Msgf("%3d scripts are processed", len(scripts))
		return p.Scripts.UpdateProjectID(scripts)
	}

	return nil
}

// Chunk -
func (p *ProjectsHandler) Chunk(lastID, size int64) ([]models.Model, error) {
	scripts, err := getScripts(p.StorageDB.DB, lastID, size)
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(scripts))
	for i := range scripts {
		data[i] = &scripts[i]
	}
	return data, nil
}

func (p *ProjectsHandler) process(script *contract.Script, chunk []contract.Script) error {
	if script.ProjectID.Valid {
		return nil
	}

	return metrics.SetScriptProjectID(p.Scripts, script, chunk)
}
