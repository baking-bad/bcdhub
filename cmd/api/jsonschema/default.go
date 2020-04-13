package jsonschema

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type defaultMaker struct{}

func (m *defaultMaker) Do(binPath string, metadata meta.Metadata) (Schema, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, fmt.Errorf("[defaultMaker] Unknown metadata binPath: %s", binPath)
	}

	schema := Schema{
		"x-props": Schema{
			"outlined": true,
			"dense":    true,
		},
	}
	switch nm.Prim {
	case consts.INT, consts.NAT, consts.MUTEZ, consts.BIGMAP:
		schema["type"] = "integer"
	case consts.STRING, consts.BYTES, consts.KEYHASH, consts.KEY, consts.ADDRESS, consts.CHAINID, consts.SIGNATURE, consts.CONTRACT, consts.LAMBDA:
		schema["type"] = "string"
	case consts.BOOL:
		schema["type"] = "boolean"
	case consts.TIMESTAMP:
		schema["type"] = "string"
		schema["format"] = "date-time"
	case consts.OPTION:
		return Create(binPath+"/o", metadata)
	default:
		return nil, fmt.Errorf("[defaultMaker] Unknown primitive: %s", nm.Prim)
	}
	if nm.Name != "" {
		schema["title"] = nm.Name
	} else {
		schema["title"] = nm.Prim
	}

	return Schema{
		"type": "object",
		"properties": Schema{
			binPath: schema,
		},
	}, nil
}
