package models

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
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
	DocBigMapActions = "bigmapaction"
	DocBigMapDiff    = "bigmapdiff"
	DocBlocks        = "block"
	DocContracts     = "contract"
	DocMigrations    = "migration"
	DocOperations    = "operation"
	DocProtocol      = "protocol"
	DocTezosDomains  = "tezos_domain"
	DocTokenBalances = "token_balance"
	DocTokenMetadata = "token_metadata"
	DocTransfers     = "transfer"
	DocTZIP          = "tzip"
)

// AllDocuments - returns all document names
func AllDocuments() []string {
	return []string{
		DocBigMapActions,
		DocBigMapDiff,
		DocBlocks,
		DocContracts,
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
		&bigmapaction.BigMapAction{},
		&bigmapdiff.BigMapDiff{},
		&block.Block{},
		&contract.Contract{},
		&migration.Migration{},
		&operation.Operation{},
		&protocol.Protocol{},
		&tezosdomain.TezosDomain{},
		&tokenbalance.TokenBalance{},
		&tokenmetadata.TokenMetadata{},
		&transfer.Transfer{},
		&tzip.TZIP{},
	}
}
