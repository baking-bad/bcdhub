package models

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/service"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Document names
const (
	DocBigMaps       = "big_maps"
	DocBigMapActions = "big_map_actions"
	DocBigMapDiff    = "big_map_diffs"
	DocBigMapState   = "big_map_states"
	DocBlocks        = "blocks"
	DocContracts     = "contracts"
	DocDApps         = "dapps"
	DocMigrations    = "migrations"
	DocOperations    = "operations"
	DocProtocol      = "protocols"
	DocServices      = "states"
	DocTezosDomains  = "tezos_domains"
	DocTokenBalances = "token_balances"
	DocTokenMetadata = "token_metadata"
	DocTransfers     = "transfers"
	DocTZIP          = "tzips"
)

// AllDocuments - returns all document names
func AllDocuments() []string {
	return []string{
		DocServices,
		DocBigMaps,
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
		&service.State{},
		&protocol.Protocol{},
		&block.Block{},
		&operation.Operation{},
		&bigmap.BigMap{},
		&bigmap.Action{},
		&bigmap.Diff{},
		&bigmap.State{},
		&contract.Contract{},
		&migration.Migration{},
		&tezosdomain.TezosDomain{},
		&tokenbalance.TokenBalance{},
		&tokenmetadata.TokenMetadata{},
		&transfer.Transfer{},
		&tzip.TZIP{},
		&dapp.DApp{},
	}
}
