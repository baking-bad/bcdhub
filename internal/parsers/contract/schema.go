package contract

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// SchemaParser -
type SchemaParser struct {
	symLink string
}

// NewSchemaParser -
func NewSchemaParser(symLink string) SchemaParser {
	return SchemaParser{symLink}
}

// Parse -
func (p SchemaParser) Parse(script gjson.Result, address string) (s schema.Schema, err error) {
	s.ID = address
	s.Storage, err = p.getSchema(script, consts.STORAGE, address)
	if err != nil {
		return
	}
	s.Parameter, err = p.getSchema(script, consts.PARAMETER, address)
	if err != nil {
		return
	}
	return
}

func (p SchemaParser) getSchema(script gjson.Result, tag, address string) (map[string]string, error) {
	res := make(map[string]string)
	schema, err := p.createSchemaSection(script, tag, address)
	if err != nil {
		return nil, err
	}
	res[p.symLink] = schema
	return res, nil
}

// Update -
func (p SchemaParser) Update(script gjson.Result, address string, s *schema.Schema) error {
	storage, err := p.createSchemaSection(script, consts.STORAGE, address)
	if err != nil {
		return err
	}
	parameter, err := p.createSchemaSection(script, consts.PARAMETER, address)
	if err != nil {
		return err
	}
	s.Storage[p.symLink] = storage
	s.Parameter[p.symLink] = parameter

	return nil
}

func (p SchemaParser) createSchemaSection(script gjson.Result, tag, address string) (string, error) {
	args := script.Get(fmt.Sprintf("code.#(prim==\"%s\").args", tag))
	if args.Exists() {
		schema, err := meta.ParseMetadata(args)
		if err != nil {
			return "", nil
		}

		b, err := json.Marshal(schema)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", errors.Errorf("[createSchemaSection] Unknown tag '%s' contract %s", tag, address)
}
