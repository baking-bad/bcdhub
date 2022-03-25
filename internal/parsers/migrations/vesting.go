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
func (p *VestingParser) Parse(data noderpc.ContractData, head noderpc.Header, address string, proto protocol.Protocol, store parsers.Store) error {
	parser := contract.NewParser(p.ctx)
	if err := parser.Parse(&operation.Operation{
		ProtocolID: proto.ID,
		Status:     types.OperationStatusApplied,
		Kind:       types.OperationKindOrigination,
		Amount:     data.Balance,
		Counter:    data.Counter,
		Source: account.Account{
			Address: data.Manager,
			Type:    types.NewAccountType(data.Manager),
		},
		Destination: account.Account{
			Address: address,
			Type:    types.NewAccountType(address),
		},
		Delegate: account.Account{
			Address: data.Delegate.Value,
			Type:    types.NewAccountType(data.Delegate.Value),
		},
		Level:     head.Level,
		Timestamp: head.Timestamp,
		Script:    data.RawScript,
	}, proto.SymLink, store); err != nil {
		return err
	}

	contracts := store.ListContracts()
	for i := range contracts {
		if contracts[i].Account.Address == address {
			store.AddMigrations(&migration.Migration{
				Level:      head.Level,
				ProtocolID: proto.ID,
				Timestamp:  head.Timestamp,
				Kind:       types.MigrationKindBootstrap,
				Contract:   contracts[i],
			})
			break
		}
	}

	return nil
}
