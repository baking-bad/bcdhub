package storage

import (
	"context"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

const (
	httpTimeout = time.Second
	ipfsTimeout = 10 * time.Second
)

// Storage -
type Storage interface {
	Get(ctx context.Context, value string, output interface{}) error
}

// Full -
type Full struct {
	bmdRepo      bigmapdiff.Repository
	contractRepo contract.Repository
	blockRepo    block.Repository
	storage      models.GeneralRepository

	rpc  noderpc.INode
	ipfs []string
}

// NewFull -
func NewFull(bmdRepo bigmapdiff.Repository, contractRepo contract.Repository, blockRepo block.Repository, storage models.GeneralRepository, rpc noderpc.INode, ipfs ...string) *Full {
	return &Full{
		bmdRepo, contractRepo, blockRepo, storage, rpc, ipfs,
	}
}

// Get -
func (f Full) Get(ctx context.Context, network types.Network, address, url string, ptr int64, output interface{}) error {
	var store Storage
	switch {
	case strings.HasPrefix(url, PrefixHTTPS), strings.HasPrefix(url, PrefixHTTP):
		store = NewHTTPStorage(
			WithTimeoutHTTP(httpTimeout),
		)
	case strings.HasPrefix(url, PrefixIPFS):
		store = NewIPFSStorage(
			f.ipfs,
			WithTimeoutIPFS(ipfsTimeout),
		)
	case strings.HasPrefix(url, PrefixSHA256):
		store = NewSha256Storage(
			WithTimeoutSha256(httpTimeout),
			WithHashSha256(url),
		)
	case strings.HasPrefix(url, PrefixTezosStorage):
		store = NewTezosStorage(f.bmdRepo, f.blockRepo, f.contractRepo, f.storage, f.rpc, address, network, ptr)
	default:
		return errors.Wrap(ErrUnknownStorageType, url)
	}

	return store.Get(ctx, url, output)
}
