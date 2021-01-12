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
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Document names
const (
	DocBalanceUpdates = "balance_update"
	DocBigMapActions  = "bigmapaction"
	DocBigMapDiff     = "bigmapdiff"
	DocBlocks         = "block"
	DocContracts      = "contract"
	DocMigrations     = "migration"
	DocOperations     = "operation"
	DocProtocol       = "protocol"
	DocSchema         = "schema"
	DocTezosDomains   = "tezos_domain"
	DocTokenBalances  = "token_balance"
	DocTransfers      = "transfer"
	DocTZIP           = "tzip"
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
		DocTezosDomains,
		DocTokenBalances,
		DocTransfers,
		DocTZIP,
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
		&schema.Schema{},
		&tezosdomain.TezosDomain{},
		&tokenbalance.TokenBalance{},
		&transfer.Transfer{},
		&tzip.TZIP{},
	}
}
