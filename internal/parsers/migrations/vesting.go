package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
)

// VestingParser -
type VestingParser struct {
	ctx *config.Context
}

// NewVestingParser -
func NewVestingParser(ctx *config.Context) *VestingParser {
	return &VestingParser{
		ctx: ctx,
	}
}

// Parse -
func (p *VestingParser) Parse(data noderpc.ContractData, head noderpc.Header, network types.Network, address string) (*parsers.Result, error) {
	proto, err := p.ctx.Cache.ProtocolByHash(network, head.Protocol)
	if err != nil {
		return nil, err
	}

	migration := &migration.Migration{
		Level:      head.Level,
		Network:    network,
		ProtocolID: proto.ID,
		Address:    address,
		Timestamp:  head.Timestamp,
		Kind:       types.MigrationKindBootstrap,
	}

	parser := contract.NewParser(p.ctx)
	contractModels, err := parser.Parse(&operation.Operation{
		Network:     network,
		ProtocolID:  proto.ID,
		Status:      types.OperationStatusApplied,
		Kind:        types.OperationKindOrigination,
		Amount:      data.Balance,
		Counter:     data.Counter,
		Source:      data.Manager,
		Destination: address,
		Delegate:    data.Delegate.Value,
		Level:       head.Level,
		Timestamp:   head.Timestamp,
		Script:      data.RawScript,
	})
	if err != nil {
		return nil, err
	}
	contractModels.Migrations = append(contractModels.Migrations, migration)
	return contractModels, nil
}
