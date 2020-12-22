package core

import (
	"io"

	"github.com/baking-bad/bcdhub/internal/models"
)

// CreateAWSRepository -
func (g *Genji) CreateAWSRepository(name, awsBucketName, awsRegion string) error {
	return nil
}

// ListRepositories -
func (g *Genji) ListRepositories() ([]models.Repository, error) {
	return nil, nil
}

// CreateSnapshots -
func (g *Genji) CreateSnapshots(repository, snapshot string, indices []string) error {
	return nil
}

// RestoreSnapshots -
func (g *Genji) RestoreSnapshots(repository, snapshot string, indices []string) error {
	return nil
}

// ListSnapshots -
func (g *Genji) ListSnapshots(repository string) (string, error) {
	return "", nil
}

// SetSnapshotPolicy -
func (g *Genji) SetSnapshotPolicy(policyID, cronSchedule, name, repository string, expireAfterInDays int64) error {
	return nil
}

// GetAllPolicies -
func (g *Genji) GetAllPolicies() ([]string, error) {
	return nil, nil
}

// GetMappings -
func (g *Genji) GetMappings(indices []string) (map[string]string, error) {
	return nil, nil
}

// CreateMapping -
func (g *Genji) CreateMapping(index string, r io.Reader) error {
	return nil
}

// ReloadSecureSettings -
func (g *Genji) ReloadSecureSettings() error {
	return nil
}
