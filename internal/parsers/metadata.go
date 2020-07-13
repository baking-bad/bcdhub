package parsers

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func getMetadata(script gjson.Result, tag, protoSymLink string, c *models.Contract) (map[string]string, error) {
	res := make(map[string]string)
	metadata, err := createMetadataSection(script, tag, c)
	if err != nil {
		return nil, err
	}
	res[protoSymLink] = metadata
	return res, nil
}

func updateMetadata(script gjson.Result, protoSymLink string, c *models.Contract, metadata *models.Metadata) error {
	storage, err := createMetadataSection(script, consts.STORAGE, c)
	if err != nil {
		return err
	}
	parameter, err := createMetadataSection(script, consts.PARAMETER, c)
	if err != nil {
		return err
	}
	metadata.Storage[protoSymLink] = storage
	metadata.Parameter[protoSymLink] = parameter

	return nil
}

func createMetadataSection(script gjson.Result, tag string, c *models.Contract) (string, error) {
	args := script.Get(fmt.Sprintf("code.#(prim==\"%s\").args", tag))
	if args.Exists() {
		metadata, err := meta.ParseMetadata(args)
		if err != nil {
			return "", nil
		}

		b, err := json.Marshal(metadata)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", fmt.Errorf("[createMetadata] Unknown tag '%s' contract %s", tag, c.Address)
}

func createMetadata(script gjson.Result, protoSymLink string, c *models.Contract) (*models.Metadata, error) {
	storage, err := getMetadata(script, consts.STORAGE, protoSymLink, c)
	if err != nil {
		return nil, err
	}
	parameter, err := getMetadata(script, consts.PARAMETER, protoSymLink, c)
	if err != nil {
		return nil, err
	}

	return &models.Metadata{
		ID:        c.Address,
		Storage:   storage,
		Parameter: parameter,
	}, nil
}
