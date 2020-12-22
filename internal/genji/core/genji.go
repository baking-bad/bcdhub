package core

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/genjidb/genji"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Genji -
type Genji struct {
	*genji.DB
}

// New -
func New() (*Genji, error) {
	db, err := genji.Open(":memory:")
	if err != nil {
		return nil, err
	}
	return &Genji{db}, nil
}

// WaitNew -
func WaitNew(timeout int) *Genji {
	var db *Genji
	var err error

	for db == nil {
		db, err = New()
		if err != nil {
			logger.Warning("Waiting elastic up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}
	return db
}

func (g *Genji) createIndexIfNotExists(index string) error {
	// TODO: check existance
	return g.Exec(fmt.Sprintf("CREATE TABLE %s", index))
}

// CreateIndexes -
func (g *Genji) CreateIndexes() error {
	for _, index := range models.AllDocuments() {
		if err := g.createIndexIfNotExists(index); err != nil {
			return err
		}
	}
	return nil
}

// DeleteByLevelAndNetwork -
func (g *Genji) DeleteByLevelAndNetwork(indices []string, network string, maxLevel int64) error {
	builder := NewBuilder()
	for i := range indices {
		builder.Delete(indices[i]).And(
			NewGt("level", maxLevel),
			NewEq("network", network),
		).Next()
	}
	return g.Exec(builder.String())
}

// DeleteIndices -
func (g *Genji) DeleteIndices(indices []string) error {
	builder := NewBuilder()
	for i := range indices {
		builder.Drop(indices[i]).Next()
	}
	return g.Exec(builder.String())
}

// DeleteByContract -
func (g *Genji) DeleteByContract(indices []string, network, address string) error {
	builder := NewBuilder()

	for i := range indices {
		builder.Delete(indices[i]).And(
			NewEq("network", network),
			NewEq("contract", address),
		).Next()
	}

	return g.Exec(builder.String())
}
