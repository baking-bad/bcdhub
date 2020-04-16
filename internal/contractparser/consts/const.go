package consts

// Keys
const (
	KeyPrim   = "prim"
	KeyArgs   = "args"
	KeyAnnots = "annots"
	KeyString = "string"
	KeyMutez  = "mutez"
	KeyBytes  = "bytes"
	KeyInt    = "int"
	KeyTime   = "timestamp"
)

//Kind
const (
	Transaction = "transaction"
	Origination = "origination"
	Delegation  = "delegation"
	Migration   = "migration"
)

// Error IDs
const (
	BadParameterError = "michelson_v1.bad_contract_parameter"
	GasExhaustedError = "gas_exhausted.operation"
)

// Statuses
const (
	Applied     = "applied"
	Backtracked = "backtracked"
	Failed      = "failed"
	Skipped     = "skipped"
)

// MigrationKind -
const (
	MigrationBootstrap = "bootstrap"
	MigrationLambda    = "lambda"
	MigrationUpdate    = "update"
)
