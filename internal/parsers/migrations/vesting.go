package migrations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
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
	ctx            *config.Context
	filesDirectory string
}

// NewVestingParser -
func NewVestingParser(ctx *config.Context, filesDirectory string) *VestingParser {
	return &VestingParser{
		ctx:            ctx,
		filesDirectory: filesDirectory,
	}
}

// Parse -
func (p *VestingParser) Parse(data noderpc.ContractData, head noderpc.Header, network types.Network, address string) (*parsers.Result, error) {
	migration := &migration.Migration{
		Level:     head.Level,
		Network:   network,
		Protocol:  head.Protocol,
		Address:   address,
		Timestamp: head.Timestamp,
		Kind:      consts.MigrationBootstrap,
	}

	op := operation.Operation{
		Network:     network,
		Protocol:    head.Protocol,
		Status:      consts.Applied,
		Kind:        consts.Migration,
		Amount:      data.Balance,
		Counter:     data.Counter,
		Source:      data.Manager,
		Destination: address,
		Delegate:    data.Delegate.Value,
		Level:       head.Level,
		Timestamp:   head.Timestamp,
		Script:      data.RawScript,
	}

	parser := contract.NewParser(p.ctx, contract.WithShareDir(p.filesDirectory))
	contractModels, err := parser.Parse(&op)
	if err != nil {
		return nil, err
	}
	contractModels.Migrations = append(contractModels.Migrations, migration)
	return contractModels, nil
}
