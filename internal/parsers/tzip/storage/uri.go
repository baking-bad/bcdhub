package storage

import (
	"net/url"
	"strings"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/pkg/errors"
)

const (
	tezosStoragePrefix = "tezos-storage:"
	sha256Prefix       = "sha256://"
)

// TezosStorageURI -
type TezosStorageURI struct {
	Address string
	Network string
	Key     string
}

// Parse -
func (uri *TezosStorageURI) Parse(value string) (err error) {
	if !strings.HasPrefix(value, tezosStoragePrefix) {
		return errors.Wrap(ErrInvalidTezosStoragePrefix, value)
	}

	uri.Key = strings.TrimPrefix(value, tezosStoragePrefix)
	if strings.HasPrefix(uri.Key, "//") {
		uri.Key = strings.TrimPrefix(value, "//")
		parts := strings.Split(uri.Key, "/")
		if len(parts) > 1 {
			uri.parseHost(parts[0])

			if len(parts) == 2 {
				uri.Key, err = url.QueryUnescape(parts[1])
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (uri *TezosStorageURI) parseHost(host string) {
	parts := strings.Split(host, ".")
	if helpers.IsAddress(parts[0]) {
		uri.Address = parts[0]
	}

	if len(parts) == 2 {
		uri.Network = parts[1]
	}
}

func (uri *TezosStorageURI) networkByChainID(blockRepo block.Repository) error {
	if uri.Network == "" {
		return nil
	}

	network, err := blockRepo.GetNetworkAlias(uri.Network)
	if err != nil {
		return err
	}
	uri.Network = network
	return nil
}

// Sha256URI -
type Sha256URI struct {
	Hash string
	Link string
}

// Parse -
func (uri *Sha256URI) Parse(value string) error {
	if !strings.HasPrefix(value, sha256Prefix) {
		return errors.Wrap(ErrInvalidSha256Prefix, value)
	}

	key := strings.TrimPrefix(value, sha256Prefix)
	parts := strings.Split(key, "/")
	if len(parts) != 2 {
		return errors.Wrap(ErrInvalidURI, value)
	}

	uri.Hash = parts[0]
	link, err := url.QueryUnescape(parts[1])
	if err != nil {
		return err
	}
	uri.Link = link
	return nil
}
