package parsers

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/models"
)

// DefaultParserOption -
type DefaultParserOption func(*DefaultParser)

// WithIPFSGateways -
func WithIPFSGateways(ipfs []string) DefaultParserOption {
	return func(dp *DefaultParser) {
		dp.ipfs = ipfs
	}
}

// WithConstants -
func WithConstants(constants models.Constants) DefaultParserOption {
	return func(dp *DefaultParser) {
		dp.constants = constants
	}
}

// WithInterfaces -
func WithInterfaces(interfaces map[string]kinds.ContractKind) DefaultParserOption {
	return func(dp *DefaultParser) {
		dp.interfaces = interfaces
	}
}

// WithTokenViews -
func WithTokenViews(views TokenViews) DefaultParserOption {
	return func(dp *DefaultParser) {
		dp.views = views
	}
}
