package main

import (
	"encoding/json"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

func getMetadata(rpc noderpc.Pool, c *models.Contract, tag, filesDirectory string) (map[string]string, error) {
	res := make(map[string]string)

	a, err := createMetadata(rpc, 0, c, tag, filesDirectory)
	if err != nil {
		return nil, err
	}

	if c.Network == consts.Mainnet {
		res[consts.MetadataBabylon] = a

		if c.Level < consts.LevelBabylon {
			a, err = createMetadata(rpc, consts.LevelBabylon-1, c, tag, filesDirectory)
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

func createMetadata(rpc noderpc.Pool, level int64, c *models.Contract, tag, filesDirectory string) (string, error) {
	s, err := contractparser.GetContract(rpc, c.Address, c.Network, level, filesDirectory)
	if err != nil {
		return "", err
	}
	script := s.Get("script")

	args := script.Get(fmt.Sprintf("code.#(prim==\"%s\").args", tag))
	if args.Exists() {
		metadata, err := meta.ParseMetadata(args)
		if err != nil {
			return "", nil
		}
		if tag == consts.PARAMETER && level == 0 {
			entrypoints, err := metadata.GetEntrypoints()
			if err != nil {
				return "", err
			}
			c.Entrypoints = make([]string, len(entrypoints))
			for i := range entrypoints {
				c.Entrypoints[i] = entrypoints[i].Name
			}
		}

		b, err := json.Marshal(metadata)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", fmt.Errorf("[createMetadata] Unknown tag '%s'", tag)
}

func saveMetadata(es *elastic.Elastic, rpc noderpc.Pool, c *models.Contract, filesDirectory string) error {
	storage, err := getMetadata(rpc, c, consts.STORAGE, filesDirectory)
	if err != nil {
		return err
	}
	parameter, err := getMetadata(rpc, c, consts.PARAMETER, filesDirectory)
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
