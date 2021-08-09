package storage

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/karlseguin/ccache"
)

// IPFS storage prefix
const (
	PrefixIPFS = "ipfs"
)

var (
	regMultihash = regexp.MustCompile("Qm[1-9A-HJ-NP-Za-km-z]{44}")
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

	multihash := strings.TrimPrefix(value, "ipfs://")
	if len(multihash) != 46 || !regMultihash.MatchString(multihash) {
		return ErrInvalidIPFSHash
	}

	if item := s.cache.Get(multihash); item != nil && !item.Expired() {
		output = item.Value()
		if output == nil {
			return ErrNoIPFSResponse
		}
		logger.Info().Str("url", value).Msg("using cached response")
		return nil
	}

	for i := range s.gateways {
		url := fmt.Sprintf("%s/ipfs/%s", s.gateways[i], multihash)
		err := s.HTTPStorage.Get(url, output)
		if err == nil {
			s.cache.Set(multihash, output, 30*24*time.Hour)
			return nil
		}
		logger.Warning().Err(err).Str("url", url).Msg("")
	}

	s.cache.Set(multihash, nil, time.Minute*5)
	return ErrNoIPFSResponse
}
