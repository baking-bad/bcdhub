package jsonschema

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type maker interface {
	Do(string, meta.Metadata) (Schema, DefaultModel, error)
}

// Schema -
type Schema map[string]interface{}

// DefaultModel -
type DefaultModel map[string]interface{}

// Extend -
func (model DefaultModel) Extend(another DefaultModel, binPath string) {
	if !strings.HasSuffix(binPath, "/o") {
		for k, v := range another {
			model[k] = v
		}
	} else {
		optionMap := make(DefaultModel)
		for k, v := range another {
			optionMap[k] = v
		}
		model[binPath] = optionMap
	}
}
