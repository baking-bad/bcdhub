package main

import (
	"encoding/json"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
)

func getMetadata(es *elastic.Elastic, address, tag string, level int64) (meta.Metadata, error) {
	if address == "" {
		return nil, fmt.Errorf("[getMetadata] Empty address")
	}

	data, err := es.GetByID(elastic.DocMetadata, address)
	if err != nil {
		return nil, err
	}
	network := meta.GetMetadataNetwork(level)
	path := fmt.Sprintf("_source.%s.%s", tag, network)
	metadata := data.Get(path).String()

	var res meta.Metadata
	if err := json.Unmarshal([]byte(metadata), &res); err != nil {
		return nil, err
	}
	return res, nil
}
