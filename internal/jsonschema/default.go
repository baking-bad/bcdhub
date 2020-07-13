package jsonschema

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type defaultMaker struct{}

func (m *defaultMaker) Do(binPath string, metadata meta.Metadata) (Schema, DefaultModel, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, nil, fmt.Errorf("[defaultMaker] Unknown metadata binPath: %s", binPath)
	}

	schema := Schema{
		"prim": nm.Prim,
	}

	model := make(DefaultModel)
	switch nm.Prim {
	case consts.INT, consts.NAT, consts.MUTEZ, consts.BIGMAP:
		schema["type"] = "integer"
	case consts.STRING, consts.BYTES, consts.KEYHASH, consts.KEY, consts.CHAINID, consts.SIGNATURE, consts.CONTRACT, consts.LAMBDA:
		schema["type"] = "string"
	case consts.BOOL:
		schema["type"] = "boolean"
		model[binPath] = false
	case consts.TIMESTAMP:
		schema["type"] = "string"
		schema["format"] = "date-time"
	case consts.ADDRESS:
		schema["type"] = "string"
		schema["minLength"] = 36
		schema["maxLength"] = 36
	case consts.OPTION:
		return Create(binPath+"/o", metadata)
	default:
		return nil, nil, fmt.Errorf("[defaultMaker] Unknown primitive: %s", nm.Prim)
	}
	if nm.Name != "" {
		schema["title"] = nm.Name
	}

	return Schema{
		"type": "object",
		"properties": Schema{
			binPath: schema,
		},
	}, model, nil
}
