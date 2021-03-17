package reindexer

import (
	"io"

	"github.com/baking-bad/bcdhub/internal/models"
)

// CreateAWSRepository -
func (r *Reindexer) CreateAWSRepository(name, awsBucketName, awsRegion string) error {
	return nil
}

// ListRepositories -
func (r *Reindexer) ListRepositories() ([]models.Repository, error) {
	return nil, nil
}

// CreateSnapshots -
func (r *Reindexer) CreateSnapshots(repository, snapshot string, indices []string) error {
	return nil
}

// RestoreSnapshots -
func (r *Reindexer) RestoreSnapshots(repository, snapshot string, indices []string) error {
	return nil
}

// ListSnapshots -
func (r *Reindexer) ListSnapshots(repository string) (string, error) {
	return "", nil
}

// SetSnapshotPolicy -
func (r *Reindexer) SetSnapshotPolicy(policyID, cronSchedule, name, repository string, expireAfterInDays int64) error {
	return nil
}

// GetAllPolicies -
func (r *Reindexer) GetAllPolicies() ([]string, error) {
	return nil, nil
}

// GetMappings -
func (r *Reindexer) GetMappings(indices []string) (map[string]string, error) {
	return nil, nil
}

// CreateMapping -
func (r *Reindexer) CreateMapping(index string, reader io.Reader) error {
	return nil
}

// ReloadSecureSettings -
func (r *Reindexer) ReloadSecureSettings() error {
	return nil
}
