package storage

import (
	"fmt"
	"math/rand"
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

	rand.Seed(time.Now().Unix())
	gateway := s.gateways[rand.Intn(len(s.gateways))]

	url := fmt.Sprintf("%s/ipfs/%s", gateway, strings.TrimPrefix(value, "ipfs://"))
	return s.HTTPStorage.Get(url, output)
}
