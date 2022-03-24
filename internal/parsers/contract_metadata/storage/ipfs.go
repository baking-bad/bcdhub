package storage

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
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
	cache *ccache.Cache
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

var globalGateways = make(map[string]time.Time)

// NewIPFSStorage -
func NewIPFSStorage(gateways []string, opts ...IPFSStorageOption) IPFSStorage {
	s := IPFSStorage{
		HTTPStorage: NewHTTPStorage(),
		cache:       ccache.New(ccache.Configure()),
	}

	if len(globalGateways) == 0 {
		if len(gateways) > 1 {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(gateways), func(i, j int) { gateways[i], gateways[j] = gateways[j], gateways[i] })
		}

		for i := range gateways {
			globalGateways[gateways[i]] = time.Now()
		}
	}

	for i := range opts {
		opts[i](&s)
	}

	return s
}

// Get -
func (s IPFSStorage) Get(ctx context.Context, value string, output interface{}) error {
	if len(globalGateways) == 0 {
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

	for baseURL, blockTime := range globalGateways {
		if time.Now().Before(blockTime) {
			continue
		}
		url := fmt.Sprintf("%s/ipfs/%s%s", baseURL, ipfsURI.Host, ipfsURI.Path)
		err := s.HTTPStorage.Get(ctx, url, output)
		if err == nil {
			s.cache.Set(value, output, MaxTTL)
			return nil
		}
		if errors.Is(err, ErrTooManyRequests) {
			globalGateways[baseURL] = time.Now().Add(3 * time.Minute)
			logger.Warning().Str("hosting", baseURL).Msg("rate limit exceeded on IPFS hosting. sleep 3 minutes")
		} else {
			logger.Warning().Err(err).Str("url", url).Msg("")
		}
	}

	s.cache.Set(value, BounceFlag, BounceTime)
	return ErrNoIPFSResponse
}
