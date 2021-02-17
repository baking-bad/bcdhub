package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
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

	return errors.Wrap(ErrNoIPFSResponse, value)
}
