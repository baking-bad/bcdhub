package tzip

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	cm "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	cmStorage "github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/storage"
)

// Public consts
const (
	EmptyStringKey = "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo"
)

type bufTzip cm.ContractMetadata

// UnmarshalJSON -
func (t *bufTzip) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*cm.ContractMetadata)(t)); err != nil {
		return err
	}
	t.Extras = make(map[string]interface{})
	if err := json.Unmarshal(data, &t.Extras); err != nil {
		return err
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
	BigMapDiff bigmapdiff.BigMapDiff
	Hash       string
}

// Parser -
type Parser struct {
	bigMapRepo   bigmapdiff.Repository
	blocksRepo   block.Repository
	contractRepo contract.Repository
	storage      models.GeneralRepository
	rpc          noderpc.INode

	cfg ParserConfig
}

// NewParser -
func NewParser(bigMapRepo bigmapdiff.Repository, blocksRepo block.Repository, contractRepo contract.Repository, storage models.GeneralRepository, rpc noderpc.INode, cfg ParserConfig) Parser {
	return Parser{
		bigMapRepo:   bigMapRepo,
		blocksRepo:   blocksRepo,
		contractRepo: contractRepo,
		storage:      storage,
		rpc:          rpc,

		cfg: cfg,
	}
}

// Parse -
func (p *Parser) Parse(ctx ParseContext) (*cm.ContractMetadata, error) {
	decoded := cmStorage.DecodeValue(ctx.BigMapDiff.Value)
	if decoded == "" {
		return nil, nil
	}

	data := new(bufTzip)
	s := cmStorage.NewFull(p.bigMapRepo, p.contractRepo, p.blocksRepo, p.storage, p.rpc, p.cfg.IPFSGateways...)
	if err := s.Get(ctx.BigMapDiff.Network, ctx.BigMapDiff.Contract, decoded, ctx.BigMapDiff.Ptr, data); err != nil {
		switch {
		case errors.Is(err, cmStorage.ErrHTTPRequest) || errors.Is(err, cmStorage.ErrJSONDecoding) || errors.Is(err, cmStorage.ErrUnknownStorageType):
			logger.Warning().Fields(ctx.BigMapDiff.LogFields()).Str("kind", "contract_metadata").Err(err).Msg("tzip.Parser.Parse")
			return nil, nil
		case errors.Is(err, cmStorage.ErrNoIPFSResponse):
			data.Description = fmt.Sprintf("Failed to fetch metadata %s", decoded)
			data.Name = consts.Unknown
			logger.Warning().Str("url", decoded).Str("kind", "contract_metadata").Err(err).Msg("")
		case p.storage.IsRecordNotFound(err):
			logger.Warning().Fields(ctx.BigMapDiff.LogFields()).Str("kind", "contract_metadata").Err(err).Msg("tzip.Parser.Parse")
			return nil, nil
		default:
			return nil, err
		}
	}
	if data == nil {
		return nil, nil
	}

	data.Address = ctx.BigMapDiff.Contract
	data.Network = ctx.BigMapDiff.Network
	data.Level = ctx.BigMapDiff.Level
	data.Timestamp = ctx.BigMapDiff.Timestamp

	return (*cm.ContractMetadata)(data), nil
}
