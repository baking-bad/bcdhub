package parsers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
)

// VestingParser -
type VestingParser struct {
	filesDirectory string
}

// NewVestingParser -
func NewVestingParser(filesDirectory string) *VestingParser {
	return &VestingParser{
		filesDirectory: filesDirectory,
	}
}

// Parse -
func (p *VestingParser) Parse(data noderpc.ContractData, head noderpc.Header, network, address string) ([]models.Model, error) {
	migration := &migration.Migration{
		IndexedTime: time.Now().UnixNano() / 1000,

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
		IndexedTime: time.Now().UnixNano() / 1000,
		Script:      data.RawScript,
	}

	parser := contract.NewParser(contract.WithShareDir(p.filesDirectory))
	contractModels, err := parser.Parse(&op)
	if err != nil {
		return nil, err
	}

	parsedModels := []models.Model{}
	if len(contractModels) > 0 {
		parsedModels = append(parsedModels, contractModels...)
	}
	parsedModels = append(parsedModels, migration)

	return parsedModels, nil
}
