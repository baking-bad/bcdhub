package main

import (
	"encoding/json"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

func getMetadata(rpc *noderpc.NodeRPC, c *models.Contract, tag string, script gjson.Result) (map[string]string, error) {
	res := make(map[string]string)

	a, err := createMetadata(rpc, 0, c, tag, &script)
	if err != nil {
		return nil, err
	}

	if c.Network == consts.Mainnet {
		res[consts.MetadataBabylon] = a

		if c.Level < consts.LevelBabylon {
			a, err = createMetadata(rpc, consts.LevelBabylon-1, c, tag, nil)
			if err != nil {
				return nil, err
			}
			res[consts.MetadataAlpha] = a
		}
	} else {
		res[consts.MetadataBabylon] = a
		res[consts.MetadataAlpha] = a
	}
	return res, nil
}

func getEntrypointsFromMetadata(m meta.Metadata, c *models.Contract) {
	root := m["0"]
	c.Entrypoints = make([]string, 0)
	if len(root.Args) > 0 {
		for i := range root.Args {
			arg := m[root.Args[i]]
			name := arg.Name
			if name == "" {
				name = fmt.Sprintf("__entry__%d", i)
			}
			c.Entrypoints = append(c.Entrypoints, name)
		}
	} else {
		c.Entrypoints = append(c.Entrypoints, "__entry__0")
	}
}

func createMetadata(rpc *noderpc.NodeRPC, level int64, c *models.Contract, tag string, script *gjson.Result) (string, error) {
	if script == nil {
		s, err := rpc.GetScriptJSON(c.Address, level)
		if err != nil {
			return "", err
		}
		script = &s
	}

	args := script.Get(fmt.Sprintf("code.#(prim==\"%s\").args", tag))
	if args.Exists() {
		a, err := meta.ParseMetadata(args)
		if err != nil {
			return "", nil
		}
		if tag == consts.PARAMETER {
			getEntrypointsFromMetadata(a, c)
		}

		b, err := json.Marshal(a)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", fmt.Errorf("[createMetadata] Unknown tag '%s'", tag)
}

func saveMetadata(es *elastic.Elastic, rpc *noderpc.NodeRPC, c *models.Contract, script gjson.Result) error {
	storage, err := getMetadata(rpc, c, consts.STORAGE, script)
	if err != nil {
		return err
	}
	parameter, err := getMetadata(rpc, c, consts.PARAMETER, script)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		consts.PARAMETER: parameter,
		consts.STORAGE:   storage,
	}
	_, err = es.AddDocumentWithID(data, elastic.DocMetadata, c.Address)
	return err
}
