package models

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Document names
const (
	DocBigMapActions = "big_map_actions"
	DocBigMapDiff    = "big_map_diffs"
	DocBigMapState   = "big_map_states"
	DocBlocks        = "blocks"
	DocContracts     = "contracts"
	DocDApps         = "dapps"
	DocMigrations    = "migrations"
	DocOperations    = "operations"
	DocProtocol      = "protocols"
	DocTezosDomains  = "tezos_domains"
	DocTokenBalances = "token_balances"
	DocTokenMetadata = "token_metadata"
	DocTransfers     = "transfers"
	DocTZIP          = "tzips"
)

// AllDocuments - returns all document names
func AllDocuments() []string {
	return []string{
		DocBigMapActions,
		DocBigMapDiff,
		DocBigMapState,
		DocBlocks,
		DocContracts,
		DocDApps,
		DocMigrations,
		DocOperations,
		DocProtocol,
		DocTezosDomains,
		DocTokenBalances,
		DocTokenMetadata,
		DocTransfers,
		DocTZIP,
	}
}

// AllModels -
func AllModels() []Model {
	return []Model{
		&protocol.Protocol{},
		&bigmapaction.BigMapAction{},
		&bigmapdiff.BigMapDiff{},
		&bigmapdiff.BigMapState{},
		&block.Block{},
		&contract.Contract{},
		&dapp.DApp{},
		&migration.Migration{},
		&operation.Operation{},
		&tezosdomain.TezosDomain{},
		&tokenbalance.TokenBalance{},
		&tokenmetadata.TokenMetadata{},
		&transfer.Transfer{},
		&tzip.TZIP{},
	}
}
