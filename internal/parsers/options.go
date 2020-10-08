package parsers

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
)

// OPGParserOption -
type OPGParserOption func(*OPGParser)

// WithIPFSGateways -
func WithIPFSGateways(ipfs []string) OPGParserOption {
	return func(dp *OPGParser) {
		dp.ipfs = ipfs
	}
}

// WithConstants -
func WithConstants(constants models.Constants) OPGParserOption {
	return func(dp *OPGParser) {
		dp.constants = constants
	}
}

// WithInterfaces -
func WithInterfaces(interfaces map[string]kinds.ContractKind) OPGParserOption {
	return func(dp *OPGParser) {
		dp.interfaces = interfaces
	}
}

// WithTokenViews -
func WithTokenViews(views transfer.TokenViews) OPGParserOption {
	return func(dp *OPGParser) {
		dp.views = views
	}
}
