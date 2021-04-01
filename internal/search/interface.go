package search

import (
	"io"

	"github.com/baking-bad/bcdhub/internal/models"
)

// Searcher -
type Searcher interface {
	ByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (Result, error)
	Save(items []Data) error
	CreateIndexes() error
	Rollback(network string, level int64) error

	CreateAWSRepository(string, string, string) error
	ListRepositories() ([]Repository, error)
	CreateSnapshots(string, string, []string) error
	RestoreSnapshots(string, string, []string) error
	ListSnapshots(string) (string, error)
	SetSnapshotPolicy(string, string, string, string, int64) error
	GetAllPolicies() ([]string, error)
	GetMappings([]string) (map[string]string, error)
	CreateMapping(string, io.Reader) error
	ReloadSecureSettings() error
}

// Data -
type Data interface {
	GetID() string
	GetIndex() string
	Prepare(model models.Model)
}
