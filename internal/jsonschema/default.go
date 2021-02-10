package jsonschema

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/pkg/errors"
)

type defaultMaker struct{}

func (m *defaultMaker) Do(binPath string, metadata meta.Metadata) (Schema, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, errors.Errorf("[defaultMaker] Unknown metadata binPath: %s", binPath)
	}

	schema := Schema{
		"prim": nm.Prim,
	}

	switch nm.Prim {
	case consts.CONTRACT:
		schema["type"] = "string" //nolint
		schema["default"] = ""

		tags, err := kinds.CheckParameterForTags(nm.Parameter)
		if err != nil {
			return nil, err
		}
		if len(tags) == 1 {
			schema["tag"] = tags[0]
		}
	case consts.INT, consts.NAT, consts.MUTEZ, consts.BIGMAP:
		schema["type"] = "integer"
		if nm.Prim != consts.BIGMAP {
			schema["default"] = 0
		}
	case consts.STRING, consts.BYTES, consts.KEYHASH, consts.KEY, consts.CHAINID, consts.SIGNATURE, consts.LAMBDA,
		consts.BAKERHASH, consts.BLS12381FR, consts.BLS12381G1, consts.BLS12381G2, consts.NEVER, consts.SAPLINGSTATE, consts.SAPLINGTRANSACTION:
		schema["type"] = "string"
		schema["default"] = ""
	case consts.BOOL:
		schema["type"] = "boolean"
		schema["default"] = false
	case consts.TIMESTAMP:
		schema["type"] = "string"
		schema["format"] = "date-time"
		schema["default"] = time.Now().UTC().Format(time.RFC3339)
	case consts.ADDRESS:
		schema["type"] = "string"
		schema["minLength"] = 36
		schema["maxLength"] = 36
		schema["default"] = ""
	case consts.TICKET:
		return Create(binPath+"/0", metadata)
	case consts.OPTION:
		return Create(binPath+"/o", metadata)
	default:
		return nil, errors.Errorf("[defaultMaker] Unknown primitive: %s", nm.Prim)
	}
	if nm.Name != "" {
		schema["title"] = nm.Name
	} else {
		schema["title"] = nm.Prim
	}

	if nm.Prim == consts.BIGMAP {
		schema["title"] = fmt.Sprintf("%s (ptr)", schema["title"])
	}

	return Schema{
		"type": "object",
		"properties": Schema{
			binPath: schema,
		},
	}, nil
}
