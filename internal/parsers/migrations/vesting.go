package migrations

import (
	"context"

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
	parser   contract.Parser
	protocol protocol.Protocol
}

// NewVestingParser -
func NewVestingParser(ctx *config.Context, contractParser contract.Parser, proto protocol.Protocol) *VestingParser {
	return &VestingParser{
		parser:   contractParser,
		protocol: proto,
	}
}

// Parse -
func (p *VestingParser) Parse(ctx context.Context, data noderpc.ContractData, head noderpc.Header, address string, store parsers.Store) error {
	if err := p.parser.Parse(ctx, &operation.Operation{
		ProtocolID: p.protocol.ID,
		Status:     types.OperationStatusApplied,
		Kind:       types.OperationKindOrigination,
		Amount:     data.Balance,
		Counter:    data.Counter,
		Source: account.Account{
			Address: data.Manager,
			Type:    types.NewAccountType(data.Manager),
			Level:   head.Level,
		},
		Destination: account.Account{
			Address: address,
			Type:    types.NewAccountType(address),
			Level:   head.Level,
		},
		Delegate: account.Account{
			Address: data.Delegate.Value,
			Type:    types.NewAccountType(data.Delegate.Value),
			Level:   head.Level,
		},
		Level:     head.Level,
		Timestamp: head.Timestamp,
		Script:    data.RawScript,
	}, store); err != nil {
		return err
	}

	contracts := store.ListContracts()
	for i := range contracts {
		if contracts[i].Account.Address == address {
			store.AddMigrations(&migration.Migration{
				Level:      head.Level,
				ProtocolID: p.protocol.ID,
				Timestamp:  head.Timestamp,
				Kind:       types.MigrationKindBootstrap,
				Contract:   contracts[i],
			})
			break
		}
	}

	return nil
}
