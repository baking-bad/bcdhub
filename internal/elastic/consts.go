package elastic

// Document names
const (
	DocContracts      = "contract"
	DocBlocks         = "block"
	DocBalanceUpdates = "balance_update"
	DocOperations     = "operation"
	DocBigMapDiff     = "bigmapdiff"
	DocBigMapActions  = "bigmapaction"
	DocMetadata       = "metadata"
	DocMigrations     = "migration"
	DocProtocol       = "protocol"
	DocTransfers      = "transfer"
	DocTZIP           = "tzip"
	DocTokenBalances  = "token_balance"
)

// Index names
const (
	IndexName = "bcd"
)

// Errors
const (
	IndexNotFoundError = "index_not_found_exception"
)
