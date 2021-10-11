package operations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
)

// ParseParams -
type ParseParams struct {
	ctx *config.Context
	rpc noderpc.INode

	constants protocol.Constants

	contractParser *contract.Parser
	transferParser *transfer.Parser

	storageParser *RichStorage

	stackTrace *stacktrace.StackTrace

	network    types.Network
	hash       string
	head       noderpc.Header
	contentIdx int64
	main       *operation.Operation
}

// ParseParamsOption -
type ParseParamsOption func(*ParseParams)

// WithConstants -
func WithConstants(constants protocol.Constants) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.constants = constants
	}
}

// WithNetwork -
func WithNetwork(network types.Network) ParseParamsOption {
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
func NewParseParams(rpc noderpc.INode, ctx *config.Context, opts ...ParseParamsOption) (*ParseParams, error) {
	params := &ParseParams{
		ctx:        ctx,
		rpc:        rpc,
		stackTrace: stacktrace.New(),
	}
	for i := range opts {
		opts[i](params)
	}

	transferParser, err := transfer.NewParser(
		rpc,
		ctx.TZIP, ctx.Blocks, ctx.TokenBalances, ctx.SharePath,
		transfer.WithStackTrace(params.stackTrace),
		transfer.WithNetwork(params.network),
		transfer.WithChainID(params.head.ChainID),
		transfer.WithGasLimit(params.constants.HardGasLimitPerOperation),
	)
	if err != nil {
		return nil, err
	}
	params.transferParser = transferParser

	params.contractParser = contract.NewParser(
		params.ctx,
		contract.WithShareDir(ctx.SharePath),
	)
	storageParser, err := NewRichStorage(ctx.BigMapDiffs, rpc, params.head.Protocol)
	if err != nil {
		return nil, err
	}
	params.storageParser = storageParser
	return params, nil
}
