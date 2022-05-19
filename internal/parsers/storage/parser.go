package storage

import (
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Parser -
type Parser interface {
	ParseTransaction(content noderpc.Operation, operation *operation.Operation, store parsers.Store) error
	ParseOrigination(content noderpc.Operation, operation *operation.Operation, store parsers.Store) error
}
