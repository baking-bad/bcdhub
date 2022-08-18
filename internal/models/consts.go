package models

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
)

// Document names
const (
	DocAccounts        = "accounts"
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
	DocScripts         = "scripts"
)

// AllDocuments - returns all document names
func AllDocuments() []string {
	return []string{
		DocAccounts,
		DocBigMapActions,
		DocBigMapDiff,
		DocBigMapState,
		DocBlocks,
		DocContracts,
		DocGlobalConstants,
		DocMigrations,
		DocOperations,
		DocProtocol,
		DocScripts,
	}
}

// AllModels -
func AllModels() []Model {
	return []Model{
		&protocol.Protocol{},
		&block.Block{},
		&account.Account{},
		&bigmapaction.BigMapAction{},
		&bigmapdiff.BigMapDiff{},
		&bigmapdiff.BigMapState{},
		&operation.Operation{},
		&contract.GlobalConstant{},
		&contract.Script{},
		&contract.ScriptConstants{},
		&contract.Contract{},
		&migration.Migration{},
	}
}

// ManyToMany -
func ManyToMany() []interface{} {
	return []interface{}{
		&contract.ScriptConstants{},
	}
}
