package tzip

import (
	"errors"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
)

// Public consts
const (
	EmptyStringKey = "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo"
)

// ParseContext -
type ParseContext struct {
	BigMapDiff bigmapdiff.BigMapDiff
	Hash       string
}

// Parser -
type Parser struct {
	bigMapRepo bigmapdiff.Repository
	blocksRepo block.Repository
	storage    models.GeneralRepository
	rpc        noderpc.INode

	cfg ParserConfig
}

// NewParser -
func NewParser(bigMapRepo bigmapdiff.Repository, blocksRepo block.Repository, storage models.GeneralRepository, rpc noderpc.INode, cfg ParserConfig) Parser {
	return Parser{
		bigMapRepo: bigMapRepo,
		blocksRepo: blocksRepo,
		storage:    storage,
		rpc:        rpc,

		cfg: cfg,
	}
}

// Parse -
func (p *Parser) Parse(ctx ParseContext) (data *tzip.TZIP, err error) {
	decoded := tzipStorage.DecodeValue(ctx.BigMapDiff.Value)
	if decoded == "" {
		return nil, nil
	}

	data = &tzip.TZIP{}
	s := tzipStorage.NewFull(p.bigMapRepo, p.blocksRepo, p.storage, p.rpc, p.cfg.SharePath, p.cfg.IPFSGateways...)
	if err := s.Get(ctx.BigMapDiff.Network, ctx.BigMapDiff.Address, decoded, ctx.BigMapDiff.Ptr, data); err != nil {
		switch {
		case errors.Is(err, tzipStorage.ErrHTTPRequest) || errors.Is(err, tzipStorage.ErrJSONDecoding):
			logger.With(&ctx.BigMapDiff).Warning(err)
			return nil, nil
		case errors.Is(err, tzipStorage.ErrNoIPFSResponse):
			data.Description = fmt.Sprintf("Failed to fetch metadata %s", decoded)
			data.Name = "Unknown"
		default:
			return nil, err
		}
	}
	if data == nil {
		return nil, nil
	}

	data.Address = ctx.BigMapDiff.Address
	data.Network = ctx.BigMapDiff.Network
	data.Level = ctx.BigMapDiff.Level
	data.Timestamp = ctx.BigMapDiff.Timestamp

	return
}
