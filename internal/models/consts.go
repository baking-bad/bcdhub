package models

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	cm "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/service"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// Document names
const (
	DocAccounts         = "accounts"
	DocBigMapActions    = "big_map_actions"
	DocBigMapDiff       = "big_map_diffs"
	DocBigMapState      = "big_map_states"
	DocBlocks           = "blocks"
	DocContracts        = "contracts"
	DocContractMetadata = "contract_metadata"
	DocDApps            = "dapps"
	DocGlobalConstants  = "global_constants"
	DocMigrations       = "migrations"
	DocOperations       = "operations"
	DocProtocol         = "protocols"
	DocServices         = "states"
	DocScripts          = "scripts"
	DocTokenBalances    = "token_balances"
	DocTokenMetadata    = "token_metadata"
	DocTransfers        = "transfers"
)

// AllDocuments - returns all document names
func AllDocuments() []string {
	return []string{
		DocServices,
		DocAccounts,
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
		DocScripts,
		DocTokenBalances,
		DocTokenMetadata,
		DocTransfers,
		DocContractMetadata,
	}
}

// AllModels -
func AllModels() []Model {
	return []Model{
		&service.State{},
		&protocol.Protocol{},
		&block.Block{},
		&account.Account{},
		&bigmapaction.BigMapAction{},
		&bigmapdiff.BigMapDiff{},
		&bigmapdiff.BigMapState{},
		&transfer.Transfer{},
		&operation.Operation{},
		&contract.GlobalConstant{},
		&contract.Script{},
		&contract.ScriptConstants{},
		&contract.Contract{},
		&migration.Migration{},
		&tokenbalance.TokenBalance{},
		&tokenmetadata.TokenMetadata{},
		&cm.ContractMetadata{},
		&dapp.DApp{},
	}
}

// ManyToMany -
func ManyToMany() []interface{} {
	return []interface{}{
		&contract.ScriptConstants{},
	}
}
