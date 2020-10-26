package contract

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// MetadataParser -
type MetadataParser struct {
	symLink string
}

// NewMetadataParser -
func NewMetadataParser(symLink string) MetadataParser {
	return MetadataParser{symLink}
}

// Parse -
func (p MetadataParser) Parse(script gjson.Result, address string) (m models.Metadata, err error) {
	m.ID = address
	m.Storage, err = p.getMetadata(script, consts.STORAGE, address)
	if err != nil {
		return
	}
	m.Parameter, err = p.getMetadata(script, consts.PARAMETER, address)
	if err != nil {
		return
	}
	return
}

func (p MetadataParser) getMetadata(script gjson.Result, tag, address string) (map[string]string, error) {
	res := make(map[string]string)
	metadata, err := p.createMetadataSection(script, tag, address)
	if err != nil {
		return nil, err
	}
	res[p.symLink] = metadata
	return res, nil
}

// UpdateMetadata -
func (p MetadataParser) UpdateMetadata(script gjson.Result, address string, metadata *models.Metadata) error {
	storage, err := p.createMetadataSection(script, consts.STORAGE, address)
	if err != nil {
		return err
	}
	parameter, err := p.createMetadataSection(script, consts.PARAMETER, address)
	if err != nil {
		return err
	}
	metadata.Storage[p.symLink] = storage
	metadata.Parameter[p.symLink] = parameter

	return nil
}

func (p MetadataParser) createMetadataSection(script gjson.Result, tag, address string) (string, error) {
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
	return "", errors.Errorf("[createMetadata] Unknown tag '%s' contract %s", tag, address)
}
