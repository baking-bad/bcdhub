package storage

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// HTTP Storage prefixes
const (
	PrefixHTTP  = "http"
	PrefixHTTPS = "https"
)

// HTTPStorage -
type HTTPStorage struct {
	timeout time.Duration
}

// HTTPStorageOption -
type HTTPStorageOption func(*HTTPStorage)

// WithTimeoutHTTP -
func WithTimeoutHTTP(timeout time.Duration) HTTPStorageOption {
	return func(s *HTTPStorage) {
		if timeout != 0 {
			s.timeout = timeout
		}
	}
}

// NewHTTPStorage -
func NewHTTPStorage(opts ...HTTPStorageOption) HTTPStorage {
	s := HTTPStorage{
		timeout: defaultTimeout,
	}

	for i := range opts {
		opts[i](&s)
	}

	return s
}

// Get -
func (s HTTPStorage) Get(value string, output interface{}) error {
	client := http.Client{
		Timeout: s.timeout,
	}
	req, err := http.NewRequest("GET", value, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(ErrHTTPRequest, err.Error())
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(output)
	return err
}
