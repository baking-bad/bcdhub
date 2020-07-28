package jsonschema

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

var makers = map[string]maker{
	"default":     &defaultMaker{},
	consts.PAIR:   &pairMaker{},
	consts.MAP:    &mapMaker{},
	consts.BIGMAP: &mapMaker{},
	consts.LIST:   &listMaker{},
	consts.SET:    &listMaker{},
	consts.OR:     &orMaker{},
}

// Create - creates json schema for metadata
func Create(binPath string, metadata meta.Metadata) (Schema, DefaultModel, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, nil, fmt.Errorf("[Create] Unknown metadata binPath: %s", binPath)
	}

	if nm.Prim == consts.UNIT {
		return nil, DefaultModel{}, nil
	}

	f, ok := makers[nm.Prim]
	if !ok {
		f = makers["default"]
	}

	schema, model, err := f.Do(binPath, metadata)
	if err != nil {
		return nil, nil, err
	}

	if strings.HasSuffix(binPath, "/o") {
		return optionWrapper(schema, binPath, metadata)
	}
	return schema, model, nil
}
