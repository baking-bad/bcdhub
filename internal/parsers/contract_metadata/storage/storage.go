package storage

import (
	"context"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
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
	ctx  *config.Context
	ipfs []string
}

// NewFull -
func NewFull(ctx *config.Context, ipfs ...string) *Full {
	return &Full{ctx, ipfs}
}

// Get -
func (f Full) Get(ctx context.Context, address, url string, ptr int64, output interface{}) error {
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
		store = NewTezosStorage(f.ctx, address, ptr)
	default:
		return errors.Wrap(ErrUnknownStorageType, url)
	}

	return store.Get(ctx, url, output)
}
