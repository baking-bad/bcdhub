package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/protocols"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
)

// ParseParams -
type ParseParams struct {
	ctx *config.Context

	specific *protocols.Specific

	stackTrace *stacktrace.StackTrace

	hash       string
	head       noderpc.Header
	contentIdx int64
	main       *operation.Operation
	protocol   *protocol.Protocol

	withEvents bool
}

// ParseParamsOption -
type ParseParamsOption func(*ParseParams)

// WithProtocol -
func WithProtocol(protocol *protocol.Protocol) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.protocol = protocol
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
func NewParseParams(ctx *config.Context, opts ...ParseParamsOption) (*ParseParams, error) {
	params := &ParseParams{
		ctx:        ctx,
		stackTrace: stacktrace.New(),
		withEvents: ctx.Network == types.Mainnet,
	}
	for i := range opts {
		opts[i](params)
	}

	if params.protocol == nil {
		proto, err := ctx.Protocols.Get(bcd.GetCurrentProtocol(), 0)
		if err != nil {
			return nil, err
		}
		params.protocol = &proto
	}

	specific, err := protocols.Get(ctx, params.protocol.Hash)
	if err != nil {
		return nil, err
	}
	params.specific = specific

	return params, nil
}
