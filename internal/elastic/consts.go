package elastic

// Document names
const (
	DocContracts  = "contract"
	DocStates     = "state"
	DocOperations = "operation"
	DocBigMapDiff = "bigmapdiff"
	DocMetadata   = "metadata"
	DocMigrations = "migration"
	DocProtocol   = "protocol"
)

// Index names
const (
	IndexName = "bcd"
)

// Errors
const (
	IndexNotFoundError = "index_not_found_exception"
	RecordNotFound     = "Record is not found:"
)
