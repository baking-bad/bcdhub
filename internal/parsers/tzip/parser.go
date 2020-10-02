package tzip

import (
	"encoding/hex"
	"strings"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
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
	Address  string
	Network  string
	Protocol string
	Hash     string
	Pointer  int64
}

// Parser -
type Parser struct {
	es  elastic.IElastic
	rpc noderpc.INode

	cfg ParserConfig

	ctx ParseContext
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
	p.ctx = ctx

	if ctx.Pointer == 0 {
		bmPtr, err := storage.FindBigMapPointer(p.es, p.rpc, ctx.Address, ctx.Network, ctx.Protocol)
		if err != nil {
			return nil, err
		}
		ctx.Pointer = bmPtr
	}

	bmd, err := p.es.GetBigMapKey(ctx.Network, EmptyStringKey, ctx.Pointer)
	if err != nil {
		return nil, err
	}
	value := gjson.Parse(bmd.Value).Get("bytes").String()
	decodedValue, err := hex.DecodeString(value)
	if err == nil {
		value = string(decodedValue)
	}
	return p.getFromStorage(value)
}

func (p Parser) getFromStorage(url string) (*models.TZIP, error) {
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
			tzipStorage.WithHashSha256(p.ctx.Hash),
		)
	case strings.HasPrefix(url, tzipStorage.PrefixTezosStorage):
		store = tzipStorage.NewTezosStorage(p.es, p.rpc, p.ctx.Address, p.ctx.Network, p.ctx.Pointer)
	default:
		return nil, errors.Wrap(ErrUnknownStorageType, url)
	}
	return store.Get(url)
}
