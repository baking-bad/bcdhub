package operations

import (
	"sync"

	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/schema"
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
	Schema        schema.Repository
	TokenBalances tokenbalance.Repository

	rpc      noderpc.INode
	shareDir string

	interfaces map[string]kinds.ContractKind
	constants  protocol.Constants

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

// WithInterfaces -
func WithInterfaces(interfaces map[string]kinds.ContractKind) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.interfaces = interfaces
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
func NewParseParams(rpc noderpc.INode, storage models.GeneralRepository, bmdRepo bigmapdiff.Repository, blockRepo block.Repository, tzipRepo tzip.Repository, schemaRepo schema.Repository, opts ...ParseParamsOption) *ParseParams {
	params := &ParseParams{
		Storage:     storage,
		BigMapDiffs: bmdRepo,
		Schema:      schemaRepo,
		rpc:         rpc,
		once:        &sync.Once{},
		stackTrace:  stacktrace.New(),
	}
	for i := range opts {
		opts[i](params)
	}

	transferParser, err := transfer.NewParser(
		params.rpc,
		tzipRepo, blockRepo, schemaRepo,
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
		params.interfaces,
		contract.WithShareDirContractParser(params.shareDir),
	)
	storageParser, err := NewRichStorage(bmdRepo, rpc, params.head.Protocol)
	if err != nil {
		logger.Error(err)
	}
	params.storageParser = storageParser
	return params
}
