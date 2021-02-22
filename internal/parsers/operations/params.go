package operations

import (
	"sync"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
)

// ParseParams -
type ParseParams struct {
	Storage       models.GeneralRepository
	BigMapDiffs   bigmapdiff.Repository
	TokenBalances tokenbalance.Repository

	rpc      noderpc.INode
	shareDir string

	constants protocol.Constants

	contractParser *contract.Parser
	transferParser *transfer.Parser

	storageParser *RichStorage

	stackTrace *stacktrace.StackTrace

	ipfs []string

	network    string
	hash       string
	head       noderpc.Header
	contentIdx int64
	main       *operation.Operation

	once *sync.Once
}

// ParseParamsOption -
type ParseParamsOption func(*ParseParams)

// WithIPFSGateways -
func WithIPFSGateways(ipfs []string) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.ipfs = ipfs
	}
}

// WithConstants -
func WithConstants(constants protocol.Constants) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.constants = constants
	}
}

// WithShareDirectory -
func WithShareDirectory(shareDir string) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.shareDir = shareDir
	}
}

// WithNetwork -
func WithNetwork(network string) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.network = network
	}
}

// WithHash -
func WithHash(hash string) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.hash = hash
	}
}

// WithHead -
func WithHead(head noderpc.Header) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.head = head
	}
}

// WithContentIndex -
func WithContentIndex(index int64) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.contentIdx = index
	}
}

// WithMainOperation -
func WithMainOperation(main *operation.Operation) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.main = main
	}
}

// NewParseParams -
func NewParseParams(rpc noderpc.INode, storage models.GeneralRepository, bmdRepo bigmapdiff.Repository, blockRepo block.Repository, tzipRepo tzip.Repository, tbRepo tokenbalance.Repository, opts ...ParseParamsOption) *ParseParams {
	params := &ParseParams{
		Storage:       storage,
		BigMapDiffs:   bmdRepo,
		TokenBalances: tbRepo,
		rpc:           rpc,
		once:          &sync.Once{},
		stackTrace:    stacktrace.New(),
	}
	for i := range opts {
		opts[i](params)
	}

	transferParser, err := transfer.NewParser(
		params.rpc,
		tzipRepo, blockRepo, storage,
		transfer.WithStackTrace(params.stackTrace),
		transfer.WithNetwork(params.network),
		transfer.WithChainID(params.head.ChainID),
		transfer.WithGasLimit(params.constants.HardGasLimitPerOperation),
	)
	if err != nil {
		logger.Error(err)
	}
	params.transferParser = transferParser

	params.contractParser = contract.NewParser(
		contract.WithShareDir(params.shareDir),
	)
	storageParser, err := NewRichStorage(bmdRepo, rpc, params.head.Protocol)
	if err != nil {
		logger.Error(err)
	}
	params.storageParser = storageParser
	return params
}
