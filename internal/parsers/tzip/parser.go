package tzip

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
)

// Public consts
const (
	EmptyStringKey = "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo"
)

type bufTzip tzip.TZIP

// UnmarshalJSON -
func (t *bufTzip) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*tzip.TZIP)(t)); err != nil {
		return err
	}
	t.Extras = make(map[string]interface{})
	if err := json.Unmarshal(data, &t.Extras); err != nil {
		return err
	}

	if _, ok := t.Extras["license"]; !ok {
		t.License = nil
	}

	for _, field := range []string{
		"name", "description", "version", "license", "homepage",
		"authors", "interfaces", "views", "events", "dapps",
	} {
		delete(t.Extras, field)
	}
	return nil
}

// ParseContext -
type ParseContext struct {
	Diff bigmap.Diff
	Hash string
}

// Parser -
type Parser struct {
	bigMapRepo bigmap.StateRepository
	blocksRepo block.Repository
	storage    models.GeneralRepository
	rpc        noderpc.INode

	cfg ParserConfig
}

// NewParser -
func NewParser(bigMapRepo bigmap.StateRepository, blocksRepo block.Repository, storage models.GeneralRepository, rpc noderpc.INode, cfg ParserConfig) Parser {
	return Parser{
		bigMapRepo: bigMapRepo,
		blocksRepo: blocksRepo,
		storage:    storage,
		rpc:        rpc,

		cfg: cfg,
	}
}

// Parse -
func (p *Parser) Parse(ctx ParseContext) (*tzip.TZIP, error) {
	decoded := tzipStorage.DecodeValue(ctx.Diff.Value)
	if decoded == "" {
		return nil, nil
	}

	data := new(bufTzip)
	s := tzipStorage.NewFull(p.bigMapRepo, p.blocksRepo, p.storage, p.rpc, p.cfg.SharePath, p.cfg.IPFSGateways...)
	if err := s.Get(ctx.Diff.BigMap.Network, ctx.Diff.BigMap.Contract, decoded, ctx.Diff.BigMap.Ptr, data); err != nil {
		switch {
		case errors.Is(err, tzipStorage.ErrHTTPRequest) || errors.Is(err, tzipStorage.ErrJSONDecoding) || errors.Is(err, tzipStorage.ErrUnknownStorageType):
			logger.Warning().Fields(ctx.Diff.LogFields()).Str("kind", "contract_metadata").Err(err).Msg("")
			return nil, nil
		case errors.Is(err, tzipStorage.ErrNoIPFSResponse):
			data.Description = fmt.Sprintf("Failed to fetch metadata %s", decoded)
			data.Name = consts.Unknown
			logger.Warning().Str("url", decoded).Str("kind", "contract_metadata").Err(err).Msg("")
		default:
			return nil, err
		}
	}
	if data == nil {
		return nil, nil
	}

	data.Address = ctx.Diff.BigMap.Contract
	data.Network = ctx.Diff.BigMap.Network
	data.Level = ctx.Diff.Level
	data.Timestamp = ctx.Diff.Timestamp

	return (*tzip.TZIP)(data), nil
}
