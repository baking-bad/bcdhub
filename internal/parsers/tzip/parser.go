package tzip

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
	"github.com/pkg/errors"
)

const (
	metadataAnnot = "metadata"
)

// Public consts
const (
	EmptyStringKey = "expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo"
)

// ParseContext -
type ParseContext struct {
	BigMapDiff models.BigMapDiff
	Hash       string
}

// Parser -
type Parser struct {
	es  elastic.IElastic
	rpc noderpc.INode

	cfg ParserConfig
}

// NewParser -
func NewParser(es elastic.IElastic, rpc noderpc.INode, cfg ParserConfig) Parser {
	return Parser{
		es:  es,
		rpc: rpc,

		cfg: cfg,
	}
}

// Parse -
func (p *Parser) Parse(ctx ParseContext) (*models.TZIP, error) {
	decoded := tzipStorage.DecodeValue(ctx.BigMapDiff.Value)
	if decoded == "" {
		return nil, nil
	}

	return p.getFromStorage(ctx, decoded)
}

func (p Parser) getFromStorage(ctx ParseContext, url string) (*models.TZIP, error) {
	var store tzipStorage.Storage
	switch {
	case strings.HasPrefix(url, tzipStorage.PrefixHTTPS), strings.HasPrefix(url, tzipStorage.PrefixHTTP):
		store = tzipStorage.NewHTTPStorage(
			tzipStorage.WithTimeoutHTTP(p.cfg.HTTPTimeout),
		)
	case strings.HasPrefix(url, tzipStorage.PrefixIPFS):
		store = tzipStorage.NewIPFSStorage(
			p.cfg.IPFSGateways,
			tzipStorage.WithTimeoutIPFS(p.cfg.HTTPTimeout),
		)
	case strings.HasPrefix(url, tzipStorage.PrefixSHA256):
		store = tzipStorage.NewSha256Storage(
			tzipStorage.WithTimeoutSha256(p.cfg.HTTPTimeout),
			tzipStorage.WithHashSha256(ctx.Hash),
		)
	case strings.HasPrefix(url, tzipStorage.PrefixTezosStorage):
		store = tzipStorage.NewTezosStorage(p.es, p.rpc, ctx.BigMapDiff.Address, ctx.BigMapDiff.Network, ctx.BigMapDiff.Ptr)
	default:
		return nil, errors.Wrap(ErrUnknownStorageType, url)
	}
	val, err := store.Get(url)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	val.Address = ctx.BigMapDiff.Address
	val.Network = ctx.BigMapDiff.Network
	val.Level = ctx.BigMapDiff.Level
	return val, nil
}
