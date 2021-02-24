package storage

import (
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

const (
	httpTimeout = time.Second
	ipfsTimeout = 10 * time.Second
)

// Storage -
type Storage interface {
	Get(value string, output interface{}) error
}

// Full -
type Full struct {
	bmdRepo   bigmapdiff.Repository
	blockRepo block.Repository
	storage   models.GeneralRepository

	rpc       noderpc.INode
	sharePath string
	ipfs      []string
}

// NewFull -
func NewFull(bmdRepo bigmapdiff.Repository, blockRepo block.Repository, storage models.GeneralRepository, rpc noderpc.INode, sharePath string, ipfs ...string) *Full {
	return &Full{
		bmdRepo, blockRepo, storage, rpc, sharePath, ipfs,
	}
}

// Get -
func (f Full) Get(network, address, url string, ptr int64, output interface{}) error {
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
		store = NewTezosStorage(f.bmdRepo, f.blockRepo, f.storage, f.rpc, address, network, f.sharePath, ptr)
	default:
		return errors.Wrap(ErrUnknownStorageType, url)
	}

	return store.Get(url, output)
}
