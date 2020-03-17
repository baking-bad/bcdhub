package main

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

func getRichStorage(es *elastic.Elastic, rpc noderpc.Pool, op gjson.Result, level int64, protocol, operationID string) (storage.RichStorage, error) {
	kind := op.Get("kind").String()

	switch protocol {
	case consts.HashBabylon, consts.HashCarthage, consts.HashZeroBabylon:
		parser := storage.NewBabylon(es, rpc)
		switch kind {
		case consts.Transaction:
			return parser.ParseTransaction(op, protocol, level, operationID)
		case consts.Origination:
			return parser.ParseOrigination(op, protocol, level, operationID)
		}
	default:
		parser := storage.NewAlpha(es)
		switch kind {
		case consts.Transaction:
			return parser.ParseTransaction(op, protocol, level, operationID)
		case consts.Origination:
			return parser.ParseOrigination(op, protocol, level, operationID)
		}
	}
	return storage.RichStorage{Empty: true}, nil
}
