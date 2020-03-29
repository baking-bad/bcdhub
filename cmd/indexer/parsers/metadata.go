package parsers

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func getMetadata(script gjson.Result, tag, protoSymLink string, c *models.Contract) (map[string]string, error) {
	res := make(map[string]string)
	metadata, err := createMetadata(script, tag, c)
	if err != nil {
		return nil, err
	}
	res[protoSymLink] = metadata
	return res, nil
}

func updateMetadata(es *elastic.Elastic, script gjson.Result, protoSymLink string, c *models.Contract) error {
	metadata, err := es.GetMetadata(c.Address)
	if err != nil {
		return err
	}
	storage, err := createMetadata(script, consts.STORAGE, c)
	if err != nil {
		return err
	}
	parameter, err := createMetadata(script, consts.PARAMETER, c)
	if err != nil {
		return err
	}

	metadata.Parameter[protoSymLink] = parameter
	metadata.Storage[protoSymLink] = storage

	_, err = es.UpdateDoc(elastic.DocMetadata, c.Address, metadata)
	return err
}

func createMetadata(script gjson.Result, tag string, c *models.Contract) (string, error) {
	args := script.Get(fmt.Sprintf("code.#(prim==\"%s\").args", tag))
	if args.Exists() {
		metadata, err := meta.ParseMetadata(args)
		if err != nil {
			return "", nil
		}
		if tag == consts.PARAMETER {
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

func saveMetadata(es *elastic.Elastic, script gjson.Result, protoSymLink string, c *models.Contract) error {
	storage, err := getMetadata(script, consts.STORAGE, protoSymLink, c)
	if err != nil {
		return err
	}
	parameter, err := getMetadata(script, consts.PARAMETER, protoSymLink, c)
	if err != nil {
		return err
	}

	upgradable, err := isUpgradable(storage[protoSymLink], parameter[protoSymLink])
	if err != nil {
		return err
	}

	if upgradable {
		c.Tags = append(c.Tags, consts.UpgradableTag)
	}

	data := map[string]interface{}{
		consts.PARAMETER: parameter,
		consts.STORAGE:   storage,
	}
	_, err = es.AddDocumentWithID(data, elastic.DocMetadata, c.Address)
	return err
}

func isUpgradable(storage, parameter string) (bool, error) {
	var paramMeta meta.Metadata
	if err := json.Unmarshal([]byte(parameter), &paramMeta); err != nil {
		return false, fmt.Errorf("Invalid parameter metadata: %v", err)
	}

	var storageMeta meta.Metadata
	if err := json.Unmarshal([]byte(storage), &storageMeta); err != nil {
		return false, fmt.Errorf("Invalid parameter metadata: %v", err)
	}

	for _, p := range paramMeta {
		if p.Type != consts.LAMBDA {
			continue
		}

		for _, s := range storageMeta {
			if s.Type != consts.LAMBDA {
				continue
			}

			if p.Parameter == s.Parameter {
				return true, nil
			}
		}
	}

	return false, nil
}
