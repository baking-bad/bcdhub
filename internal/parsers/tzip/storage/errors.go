package storage

import "errors"

// Errors
var (
	ErrInvalidTezosStoragePrefix = errors.New("invalid tezos storage prefix")
	ErrInvalidSha256Prefix       = errors.New("invalid sha256 prefix")
	ErrInvalidURI                = errors.New("invalid URI")
	ErrEmptyIPFSGatewayList      = errors.New("empty IPFS gateway list")
	ErrUnknownBigMapPointer      = errors.New("unknown big map pointer `metadata`")
	ErrUnknownStorageType        = errors.New("unknown storage type")
	ErrHTTPRequest               = errors.New("HTTP request error")
	ErrJSONDecoding              = errors.New("JSON decoding error")
	ErrNoIPFSResponse            = errors.New("can't load document from IPFS")
	ErrInvalidIPFSHash           = errors.New("Invalid IPFS multihash")
)
