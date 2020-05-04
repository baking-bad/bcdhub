package jsonschema

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type maker interface {
	Do(string, meta.Metadata) (Schema, error)
}

// Schema -
type Schema map[string]interface{}
