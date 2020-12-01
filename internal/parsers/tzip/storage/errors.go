package storage

import "errors"

// Errors
var (
	ErrInvalidTezosStoragePrefix = errors.New("Invalid tezos storage prefix")
	ErrInvalidSha256Prefix       = errors.New("Invalid sha256 prefix")
	ErrInvalidURI                = errors.New("Invalid URI")
	ErrEmptyIPFSGatewayList      = errors.New("Empty IPFS gateway list")
	ErrUnknownBigMapPointer      = errors.New("Unknown big map pointer `metadata`")
)
