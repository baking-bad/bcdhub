package operations

import (
	"sync"

	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
)

// ParseParams -
type ParseParams struct {
	rpc      noderpc.INode
	es       elastic.IElastic
	shareDir string

	interfaces map[string]kinds.ContractKind
	constants  models.Constants

	contractParser *contract.Parser
	transferParser *transfer.Parser

	storageParser *RichStorage

	ipfs   []string
	events transfer.TokenEvents

	network    string
	hash       string
	head       noderpc.Header
	contentIdx int64
	main       *models.Operation

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
func WithConstants(constants models.Constants) ParseParamsOption {
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

// WithTokenEvents -
func WithTokenEvents(events transfer.TokenEvents) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.events = events
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
func WithMainOperation(main *models.Operation) ParseParamsOption {
	return func(dp *ParseParams) {
		dp.main = main
	}
}

// NewParseParams -
func NewParseParams(rpc noderpc.INode, es elastic.IElastic, opts ...ParseParamsOption) *ParseParams {
	params := &ParseParams{
		es:   es,
		rpc:  rpc,
		once: &sync.Once{},
	}
	for i := range opts {
		opts[i](params)
	}

	params.transferParser = transfer.NewParser(
		params.rpc, params.es,
		transfer.WithTokenViews(params.events),
	)
	params.contractParser = contract.NewParser(
		params.interfaces,
		contract.WithShareDirContractParser(params.shareDir),
	)
	storageParser, err := NewRichStorage(es, rpc, params.head.Protocol)
	if err != nil {
		logger.Error(err)
	}
	params.storageParser = storageParser
	return params
}
