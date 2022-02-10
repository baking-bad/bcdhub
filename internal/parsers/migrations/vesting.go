package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
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
func (p *VestingParser) Parse(data noderpc.ContractData, head noderpc.Header, network types.Network, address string, proto protocol.Protocol, result *parsers.Result) error {
	parser := contract.NewParser(p.ctx)
	if err := parser.Parse(&operation.Operation{
		Network:    network,
		ProtocolID: proto.ID,
		Status:     types.OperationStatusApplied,
		Kind:       types.OperationKindOrigination,
		Amount:     data.Balance,
		Counter:    data.Counter,
		Source: account.Account{
			Network: network,
			Address: data.Manager,
			Type:    types.NewAccountType(data.Manager),
		},
		Destination: account.Account{
			Network: network,
			Address: address,
			Type:    types.NewAccountType(address),
		},
		Delegate: account.Account{
			Network: network,
			Address: data.Delegate.Value,
			Type:    types.NewAccountType(data.Delegate.Value),
		},
		Level:     head.Level,
		Timestamp: head.Timestamp,
		Script:    data.RawScript,
	}, proto.SymLink, result); err != nil {
		return err
	}

	for i := range result.Contracts {
		if result.Contracts[i].Network == network && result.Contracts[i].Account.Address == address {
			result.Migrations = append(result.Migrations, &migration.Migration{
				Level:      head.Level,
				ProtocolID: proto.ID,
				Timestamp:  head.Timestamp,
				Kind:       types.MigrationKindBootstrap,
				Contract:   result.Contracts[i],
			})
			break
		}
	}

	return nil
}
