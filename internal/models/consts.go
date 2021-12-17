package models

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/service"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Document names
const (
	DocBigMapActions   = "big_map_actions"
	DocBigMapDiff      = "big_map_diffs"
	DocBigMapState     = "big_map_states"
	DocBlocks          = "blocks"
	DocContracts       = "contracts"
	DocDApps           = "dapps"
	DocGlobalConstants = "global_constants"
	DocMigrations      = "migrations"
	DocOperations      = "operations"
	DocProtocol        = "protocols"
	DocServices        = "states"
	DocTokenBalances   = "token_balances"
	DocTokenMetadata   = "token_metadata"
	DocTransfers       = "transfers"
	DocTZIP            = "tzips"
)

// AllDocuments - returns all document names
func AllDocuments() []string {
	return []string{
		DocServices,
		DocBigMapActions,
		DocBigMapDiff,
		DocBigMapState,
		DocBlocks,
		DocContracts,
		DocDApps,
		DocGlobalConstants,
		DocMigrations,
		DocOperations,
		DocProtocol,
		DocTokenBalances,
		DocTokenMetadata,
		DocTransfers,
		DocTZIP,
	}
}

// AllModels -
func AllModels() []Model {
	return []Model{
		&service.State{},
		&protocol.Protocol{},
		&block.Block{},
		&bigmapaction.BigMapAction{},
		&bigmapdiff.BigMapDiff{},
		&bigmapdiff.BigMapState{},
		&transfer.Transfer{},
		&operation.Operation{},
		&global_constant.GlobalConstant{},
		&contract.ContractConstants{},
		&contract.Contract{},
		&migration.Migration{},
		&tokenbalance.TokenBalance{},
		&tokenmetadata.TokenMetadata{},
		&tzip.TZIP{},
		&dapp.DApp{},
	}
}

// ManyToMany -
func ManyToMany() []interface{} {
	return []interface{}{
		&contract.ContractConstants{},
	}
}
