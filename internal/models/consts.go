package models

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
