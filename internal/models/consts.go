package models

import (
	"github.com/baking-bad/bcdhub/internal/models/balanceupdate"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Document names
const (
	DocContracts      = "contract"
	DocBlocks         = "block"
	DocBalanceUpdates = "balance_update"
	DocOperations     = "operation"
	DocBigMapDiff     = "bigmapdiff"
	DocBigMapActions  = "bigmapaction"
	DocSchema         = "schema"
	DocMigrations     = "migration"
	DocProtocol       = "protocol"
	DocTransfers      = "transfer"
	DocTZIP           = "tzip"
	DocTokenBalances  = "token_balance"
	DocTezosDomains   = "tezos_domain"
)

// AllDocuments - returns all document names
func AllDocuments() []string {
	return []string{
		DocBalanceUpdates,
		DocBigMapActions,
		DocBigMapDiff,
		DocBlocks,
		DocContracts,
		DocMigrations,
		DocOperations,
		DocProtocol,
		DocSchema,
		DocTZIP,
		DocTezosDomains,
		DocTokenBalances,
		DocTransfers,
	}
}

// AllModels -
func AllModels() []Model {
	return []Model{
		&balanceupdate.BalanceUpdate{},
		&bigmapaction.BigMapAction{},
		&bigmapdiff.BigMapDiff{},
		&block.Block{},
		&contract.Contract{},
		&migration.Migration{},
		&operation.Operation{},
		&protocol.Protocol{},
		&transfer.Transfer{},
		&tzip.TZIP{},
		&tokenbalance.TokenBalance{},
		&tezosdomain.TezosDomain{},
	}
}
