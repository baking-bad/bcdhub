package core

import (
	"io"

	"github.com/baking-bad/bcdhub/internal/models"
)

// CreateAWSRepository -
func (p *Postgres) CreateAWSRepository(string, string, string) error {
	return nil
}

// ListRepositories -
func (p *Postgres) ListRepositories() ([]models.Repository, error) {
	return nil, nil
}

// CreateSnapshots -
func (p *Postgres) CreateSnapshots(string, string, []string) error {
	return nil
}

// RestoreSnapshots -
func (p *Postgres) RestoreSnapshots(string, string, []string) error {
	return nil
}

// ListSnapshots -
func (p *Postgres) ListSnapshots(string) (string, error) {
	return "", nil
}

// SetSnapshotPolicy -
func (p *Postgres) SetSnapshotPolicy(string, string, string, string, int64) error {
	return nil
}

// GetAllPolicies -
func (p *Postgres) GetAllPolicies() ([]string, error) {
	return nil, nil
}

// GetMappings -
func (p *Postgres) GetMappings([]string) (map[string]string, error) {
	return nil, nil
}

// CreateMapping -
func (p *Postgres) CreateMapping(string, io.Reader) error {
	return nil
}

// ReloadSecureSettings -
func (p *Postgres) ReloadSecureSettings() error {
	return nil
}
