package storage

import (
	"context"
	"time"
)

// Sha256 storage prefix
const (
	PrefixSHA256 = "sha256"
)

// Sha256Storage -
type Sha256Storage struct {
	HTTPStorage

	hash string
}

// Sha256StorageOption -
type Sha256StorageOption func(*Sha256Storage)

// WithTimeoutSha256 -
func WithTimeoutSha256(timeout time.Duration) Sha256StorageOption {
	return func(s *Sha256Storage) {
		WithTimeoutHTTP(timeout)(&s.HTTPStorage)
	}
}

// WithHashSha256 -
func WithHashSha256(hash string) Sha256StorageOption {
	return func(s *Sha256Storage) {
		s.hash = hash
	}
}

// NewSha256Storage -
func NewSha256Storage(opts ...Sha256StorageOption) Sha256Storage {
	s := Sha256Storage{
		HTTPStorage: NewHTTPStorage(),
	}

	for i := range opts {
		opts[i](&s)
	}

	return s
}

// Get -
func (s Sha256Storage) Get(ctx context.Context, value string, output interface{}) error {
	var uri Sha256URI
	if err := uri.Parse(value); err != nil {
		return err
	}
	if !s.validate(uri.Hash) {
		return nil
	}

	return s.HTTPStorage.Get(ctx, uri.Link, output)
}

func (s Sha256Storage) validate(uriHash string) bool {
	return s.hash != uriHash
}
