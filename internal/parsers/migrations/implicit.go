package migrations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
)

// ImplicitParser -
type ImplicitParser struct {
	ctx      *config.Context
	network  types.Network
	rpc      noderpc.INode
	protocol protocol.Protocol
}

// NewImplicitParser -
func NewImplicitParser(ctx *config.Context, network types.Network, rpc noderpc.INode, protocol protocol.Protocol) *ImplicitParser {
	return &ImplicitParser{ctx, network, rpc, protocol}
}

// Parse -
func (p *ImplicitParser) Parse(metadata noderpc.Metadata, head noderpc.Header) (*parsers.Result, error) {
	if len(metadata.ImplicitOperationsResults) == 0 {
		return nil, nil
	}

	parserResult := parsers.NewResult()
	for i := range metadata.ImplicitOperationsResults {
		switch metadata.ImplicitOperationsResults[i].Kind {
		case consts.Origination:
			if err := p.origination(metadata.ImplicitOperationsResults[i], head, p.protocol.ID, parserResult); err != nil {
				return nil, err
			}
		case consts.Transaction:
		}
	}
	return parserResult, nil
}

func (p *ImplicitParser) origination(implicit noderpc.ImplicitOperationsResult, head noderpc.Header, protocolID int64, result *parsers.Result) error {
	origination := operation.Operation{
		Network:    p.network,
		ProtocolID: protocolID,
		Level:      head.Level,
		Timestamp:  head.Timestamp,
		Kind:       types.NewOperationKind(implicit.Kind),
		Destination: account.Account{
			Network: p.network,
			Address: implicit.OriginatedContracts[0],
			Type:    types.AccountTypeContract,
		},
		ConsumedGas:         implicit.ConsumedGas,
		PaidStorageSizeDiff: implicit.PaidStorageSizeDiff,
		StorageSize:         implicit.StorageSize,
		DeffatedStorage:     implicit.Storage,
	}

	script, err := p.rpc.GetRawScript(origination.Destination.Address, origination.Level)
	if err != nil {
		return err
	}
	origination.Script = script

	contractParser := contract.NewParser(p.ctx)
	if err := contractParser.Parse(&origination, p.protocol.SymLink, result); err != nil {
		return err
	}

	for i := range result.Contracts {
		if result.Contracts[i].Network == p.network && result.Contracts[i].Account.Address == implicit.OriginatedContracts[0] {
			result.Migrations = append(result.Migrations, &migration.Migration{
				ProtocolID: protocolID,
				Level:      head.Level,
				Timestamp:  head.Timestamp,
				Kind:       types.MigrationKindBootstrap,
				Contract:   result.Contracts[i],
			})
			break
		}
	}

	logger.Info().Msg("Implicit bootstrap migration found")

	return nil
}
