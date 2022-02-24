package storage

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/karlseguin/ccache"
)

// IPFS storage prefix
const (
	PrefixIPFS = "ipfs"
	MaxTTL     = time.Duration(1<<63 - 1)
	BounceTime = time.Minute * 5
	BounceFlag = "no_response"
)

// IPFSStorage -
type IPFSStorage struct {
	HTTPStorage
	gateways []string
	cache    *ccache.Cache
}

// IPFSStorageOption -
type IPFSStorageOption func(*IPFSStorage)

// WithTimeoutIPFS -
func WithTimeoutIPFS(timeout time.Duration) IPFSStorageOption {
	return func(s *IPFSStorage) {
		if ipfsTimeout := os.Getenv("IPFS_TIMEOUT"); ipfsTimeout != "" {
			seconds, err := strconv.ParseInt(ipfsTimeout, 10, 64)
			if err == nil {
				WithTimeoutHTTP(time.Duration(seconds) * time.Second)(&s.HTTPStorage)
				return
			}
		}

		WithTimeoutHTTP(timeout)(&s.HTTPStorage)
	}
}

// NewIPFSStorage -
func NewIPFSStorage(gateways []string, opts ...IPFSStorageOption) IPFSStorage {
	s := IPFSStorage{
		HTTPStorage: NewHTTPStorage(),
		gateways:    gateways,
		cache:       ccache.New(ccache.Configure()),
	}

	for i := range opts {
		opts[i](&s)
	}

	return s
}

// Get -
func (s IPFSStorage) Get(value string, output interface{}) error {
	if len(s.gateways) == 0 {
		return ErrEmptyIPFSGatewayList
	}

	ipfsURI, err := url.Parse(value)
	if err != nil {
		return ErrInvalidIPFSHash
	}

	if ipfsURI.Scheme != "ipfs" {
		return ErrInvalidIPFSHash
	}

	if !helpers.IsIPFS(ipfsURI.Host) {
		return ErrInvalidIPFSHash
	}

	if item := s.cache.Get(value); item != nil && !item.Expired() {
		output = item.Value()
		logger.Info().Str("url", value).Msg("using cached response")
		if output == BounceFlag {
			return ErrNoIPFSResponse
		}
		return nil
	}

	for i := range s.gateways {
		url := fmt.Sprintf("%s/ipfs/%s%s", s.gateways[i], ipfsURI.Host, ipfsURI.Path)
		if err := s.HTTPStorage.Get(url, output); err == nil {
			s.cache.Set(value, output, MaxTTL)
			return nil
		}
		logger.Warning().Err(err).Str("url", url).Msg("")
	}

	s.cache.Set(value, BounceFlag, BounceTime)
	return ErrNoIPFSResponse
}
