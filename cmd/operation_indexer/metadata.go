package main

import (
	"encoding/json"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
)

func getMetadata(es *elastic.Elastic, address string) (map[string]contractparser.Metadata, error) {
	if address == "" {
		return nil, fmt.Errorf("[getMetadata] Empty address")
	}

	data, err := es.GetByID(elastic.DocMetadata, address)
	if err != nil {
		return nil, err
	}
	res := make(map[string]contractparser.Metadata)
	for k, v := range data.Get("_source").Map() {
		var m contractparser.Metadata
		if err := json.Unmarshal([]byte(v.String()), &m); err != nil {
			return nil, err
		}
		res[k] = m
	}

	return res, nil
}
