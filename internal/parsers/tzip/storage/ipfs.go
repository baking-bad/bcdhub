package storage

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// IPFS storage prefix
const (
	PrefixIPFS = "ipfs"
)

// IPFSStorage -
type IPFSStorage struct {
	HTTPStorage
	gateways []string
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

	for i := range s.gateways {
		url := fmt.Sprintf("%s/ipfs/%s", s.gateways[i], strings.TrimPrefix(value, "ipfs://"))
		if err := s.HTTPStorage.Get(url, output); err == nil {
			return nil
		}
	}

	return ErrNoIPFSResponse
}
