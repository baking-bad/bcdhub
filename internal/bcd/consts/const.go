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

// Kind
const (
	Transaction    = "transaction"
	Origination    = "origination"
	OriginationNew = "origination_new"
	Delegation     = "delegation"
	Migration      = "migration"
)

// Error IDs
const (
	BadParameterError             = "michelson_v1.bad_contract_parameter"
	GasExhaustedError             = "gas_exhausted.operation"
	ScriptRejectedError           = "michelson_v1.script_rejected"
	BalanceTooLowError            = "contract.balance_too_low"
	InvalidSyntacticConstantError = "invalidSyntacticConstantError"
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

// Multisig -
const (
	MultisigScript1 = `[{'prim': 'parameter', 'args': [{'prim': 'or', 'args': [{'prim': 'unit', 'annots': ['%default']}, {'prim': 'pair', 'annots': ['%main'], 'args': [{'prim': 'pair', 'annots': [':payload'], 'args': [{'prim': 'nat', 'annots': ['%counter']}, {'prim': 'or', 'annots': [':action'], 'args': [{'prim': 'lambda', 'annots': ['%operation'], 'args': [{'prim': 'unit'}, {'prim': 'operation'}]}, {'prim': 'pair', 'annots': ['%change_keys'], 'args': [{'prim': 'nat', 'annots': ['%threshold']}, {'prim': 'list', 'annots': ['%keys'], 'args': [{'prim': 'key'}]}]}]}]}, {'prim': 'list', 'annots': ['%sigs'], 'args': [{'prim': 'option', 'args': [{'prim': 'signature'}]}]}]}]}]}, {'prim': 'storage', 'args': [{'prim': 'pair', 'args': [{'prim': 'nat', 'annots': ['%stored_counter']}, {'prim': 'pair', 'args': [{'prim': 'nat', 'annots': ['%threshold']}, {'prim': 'list', 'annots': ['%keys'], 'args': [{'prim': 'key'}]}]}]}]}, {'prim': 'code', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'IF_LEFT', 'args': [[{'prim': 'DROP'}, {'prim': 'NIL', 'args': [{'prim': 'operation'}]}, {'prim': 'PAIR'}], [{'prim': 'PUSH', 'args': [{'prim': 'mutez'}, {'int': '0'}]}, {'prim': 'AMOUNT'}, [[{'prim': 'COMPARE'}, {'prim': 'EQ'}], {'prim': 'IF', 'args': [[], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}], {'prim': 'SWAP'}, {'prim': 'DUP'}, {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DUP'}, {'prim': 'SELF'}, {'prim': 'ADDRESS'}, {'prim': 'PAIR'}, {'prim': 'PACK'}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@counter']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}]]}, {'prim': 'SWAP'}]]}, [{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@stored_counter']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, [[{'prim': 'COMPARE'}, {'prim': 'EQ'}], {'prim': 'IF', 'args': [[], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, [{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@threshold']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR', 'annots': ['@keys']}]]}], {'prim': 'DIP', 'args': [[{'prim': 'PUSH', 'annots': ['@valid'], 'args': [{'prim': 'nat'}, {'int': '0'}]}, {'prim': 'SWAP'}, {'prim': 'ITER', 'args': [[{'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, {'prim': 'SWAP'}, {'prim': 'IF_CONS', 'args': [[[{'prim': 'IF_NONE', 'args': [[{'prim': 'SWAP'}, {'prim': 'DROP'}], [{'prim': 'SWAP'}, {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}, [{'prim': 'DIP', 'args': [{'int': '2'}, [[[{'prim': 'DIP', 'args': [[{'prim': 'DUP'}]]}, {'prim': 'SWAP'}]]]]}], [[{'prim': 'DIP', 'args': [{'int': '2'}, [{'prim': 'DUP'}]]}, {'prim': 'DIG', 'args': [{'int': '3'}]}], {'prim': 'DIP', 'args': [[{'prim': 'CHECK_SIGNATURE'}]]}, {'prim': 'SWAP'}, {'prim': 'IF', 'args': [[{'prim': 'DROP'}], [{'prim': 'FAILWITH'}]]}], {'prim': 'PUSH', 'args': [{'prim': 'nat'}, {'int': '1'}]}, {'prim': 'ADD', 'annots': ['@valid']}]]}]]}]], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}, {'prim': 'SWAP'}]]}]]}, [[{'prim': 'COMPARE'}, {'prim': 'LE'}], {'prim': 'IF', 'args': [[], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}], {'prim': 'IF_CONS', 'args': [[[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]], []]}, {'prim': 'DROP'}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'PUSH', 'args': [{'prim': 'nat'}, {'int': '1'}]}, {'prim': 'ADD', 'annots': ['@new_counter']}, {'prim': 'PAIR'}]]}, {'prim': 'NIL', 'args': [{'prim': 'operation'}]}, {'prim': 'SWAP'}, {'prim': 'IF_LEFT', 'args': [[{'prim': 'UNIT'}, {'prim': 'EXEC'}, {'prim': 'CONS'}], [{'prim': 'DIP', 'args': [[{'prim': 'SWAP'}, {'prim': 'CAR'}]]}, {'prim': 'SWAP'}, {'prim': 'PAIR'}, {'prim': 'SWAP'}]]}, {'prim': 'PAIR'}]]}]]}]`
	MultisigScript2 = `[{'prim': 'parameter', 'args': [{'prim': 'or', 'args': [{'prim': 'unit', 'annots': ['%default']}, {'prim': 'pair', 'annots': ['%main'], 'args': [{'prim': 'pair', 'annots': [':payload'], 'args': [{'prim': 'nat', 'annots': ['%counter']}, {'prim': 'or', 'annots': [':action'], 'args': [{'prim': 'lambda', 'annots': ['%operation'], 'args': [{'prim': 'unit'}, {'prim': 'list', 'args': [{'prim': 'operation'}]}]}, {'prim': 'pair', 'annots': ['%change_keys'], 'args': [{'prim': 'nat', 'annots': ['%threshold']}, {'prim': 'list', 'annots': ['%keys'], 'args': [{'prim': 'key'}]}]}]}]}, {'prim': 'list', 'annots': ['%sigs'], 'args': [{'prim': 'option', 'args': [{'prim': 'signature'}]}]}]}]}]}, {'prim': 'storage', 'args': [{'prim': 'pair', 'args': [{'prim': 'nat', 'annots': ['%stored_counter']}, {'prim': 'pair', 'args': [{'prim': 'nat', 'annots': ['%threshold']}, {'prim': 'list', 'annots': ['%keys'], 'args': [{'prim': 'key'}]}]}]}]}, {'prim': 'code', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'IF_LEFT', 'args': [[{'prim': 'DROP'}, {'prim': 'NIL', 'args': [{'prim': 'operation'}]}, {'prim': 'PAIR'}], [{'prim': 'PUSH', 'args': [{'prim': 'mutez'}, {'int': '0'}]}, {'prim': 'AMOUNT'}, [[{'prim': 'COMPARE'}, {'prim': 'EQ'}], {'prim': 'IF', 'args': [[], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}], {'prim': 'SWAP'}, {'prim': 'DUP'}, {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DUP'}, {'prim': 'SELF'}, {'prim': 'ADDRESS'}, {'prim': 'PAIR'}, {'prim': 'PACK'}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@counter']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}]]}, {'prim': 'SWAP'}]]}, [{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@stored_counter']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, [[{'prim': 'COMPARE'}, {'prim': 'EQ'}], {'prim': 'IF', 'args': [[], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, [{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@threshold']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR', 'annots': ['@keys']}]]}], {'prim': 'DIP', 'args': [[{'prim': 'PUSH', 'annots': ['@valid'], 'args': [{'prim': 'nat'}, {'int': '0'}]}, {'prim': 'SWAP'}, {'prim': 'ITER', 'args': [[{'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, {'prim': 'SWAP'}, {'prim': 'IF_CONS', 'args': [[[{'prim': 'IF_NONE', 'args': [[{'prim': 'SWAP'}, {'prim': 'DROP'}], [{'prim': 'SWAP'}, {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}, [{'prim': 'DIP', 'args': [{'int': '2'}, [[[{'prim': 'DIP', 'args': [[{'prim': 'DUP'}]]}, {'prim': 'SWAP'}]]]]}], [[{'prim': 'DIP', 'args': [{'int': '2'}, [{'prim': 'DUP'}]]}, {'prim': 'DIG', 'args': [{'int': '3'}]}], {'prim': 'DIP', 'args': [[{'prim': 'CHECK_SIGNATURE'}]]}, {'prim': 'SWAP'}, {'prim': 'IF', 'args': [[{'prim': 'DROP'}], [{'prim': 'FAILWITH'}]]}], {'prim': 'PUSH', 'args': [{'prim': 'nat'}, {'int': '1'}]}, {'prim': 'ADD', 'annots': ['@valid']}]]}]]}]], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}, {'prim': 'SWAP'}]]}]]}, [[{'prim': 'COMPARE'}, {'prim': 'LE'}], {'prim': 'IF', 'args': [[], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}], {'prim': 'IF_CONS', 'args': [[[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]], []]}, {'prim': 'DROP'}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'PUSH', 'args': [{'prim': 'nat'}, {'int': '1'}]}, {'prim': 'ADD', 'annots': ['@new_counter']}, {'prim': 'PAIR'}]]}, {'prim': 'IF_LEFT', 'args': [[{'prim': 'UNIT'}, {'prim': 'EXEC'}], [{'prim': 'DIP', 'args': [[{'prim': 'CAR'}]]}, {'prim': 'SWAP'}, {'prim': 'PAIR'}, {'prim': 'NIL', 'args': [{'prim': 'operation'}]}]]}, {'prim': 'PAIR'}]]}]]}]`
	MultisigScript3 = `[{'prim': 'parameter', 'args': [{'prim': 'pair', 'args': [{'prim': 'pair', 'annots': [':payload'], 'args': [{'prim': 'nat', 'annots': ['%counter']}, {'prim': 'or', 'annots': [':action'], 'args': [{'prim': 'pair', 'annots': [':transfer'], 'args': [{'prim': 'mutez', 'annots': ['%amount']}, {'prim': 'contract', 'annots': ['%dest'], 'args': [{'prim': 'unit'}]}]}, {'prim': 'or', 'args': [{'prim': 'option', 'annots': ['%delegate'], 'args': [{'prim': 'key_hash'}]}, {'prim': 'pair', 'annots': ['%change_keys'], 'args': [{'prim': 'nat', 'annots': ['%threshold']}, {'prim': 'list', 'annots': ['%keys'], 'args': [{'prim': 'key'}]}]}]}]}]}, {'prim': 'list', 'annots': ['%sigs'], 'args': [{'prim': 'option', 'args': [{'prim': 'signature'}]}]}]}]}, {'prim': 'storage', 'args': [{'prim': 'pair', 'args': [{'prim': 'nat', 'annots': ['%stored_counter']}, {'prim': 'pair', 'args': [{'prim': 'nat', 'annots': ['%threshold']}, {'prim': 'list', 'annots': ['%keys'], 'args': [{'prim': 'key'}]}]}]}]}, {'prim': 'code', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'SWAP'}, {'prim': 'DUP'}, {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DUP'}, {'prim': 'SELF'}, {'prim': 'ADDRESS'}, {'prim': 'CHAIN_ID'}, {'prim': 'PAIR'}, {'prim': 'PAIR'}, {'prim': 'PACK'}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@counter']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}]]}, {'prim': 'SWAP'}]]}, [{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@stored_counter']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, [[{'prim': 'COMPARE'}, {'prim': 'EQ'}], {'prim': 'IF', 'args': [[], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}], {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, [{'prim': 'DUP'}, {'prim': 'CAR', 'annots': ['@threshold']}, {'prim': 'DIP', 'args': [[{'prim': 'CDR', 'annots': ['@keys']}]]}], {'prim': 'DIP', 'args': [[{'prim': 'PUSH', 'annots': ['@valid'], 'args': [{'prim': 'nat'}, {'int': '0'}]}, {'prim': 'SWAP'}, {'prim': 'ITER', 'args': [[{'prim': 'DIP', 'args': [[{'prim': 'SWAP'}]]}, {'prim': 'SWAP'}, {'prim': 'IF_CONS', 'args': [[[{'prim': 'IF_NONE', 'args': [[{'prim': 'SWAP'}, {'prim': 'DROP'}], [{'prim': 'SWAP'}, {'prim': 'DIP', 'args': [[{'prim': 'SWAP'}, [{'prim': 'DIP', 'args': [{'int': '2'}, [[[{'prim': 'DIP', 'args': [[{'prim': 'DUP'}]]}, {'prim': 'SWAP'}]]]]}], [[{'prim': 'DIP', 'args': [{'int': '2'}, [{'prim': 'DUP'}]]}, {'prim': 'DIG', 'args': [{'int': '3'}]}], {'prim': 'DIP', 'args': [[{'prim': 'CHECK_SIGNATURE'}]]}, {'prim': 'SWAP'}, {'prim': 'IF', 'args': [[{'prim': 'DROP'}], [{'prim': 'FAILWITH'}]]}], {'prim': 'PUSH', 'args': [{'prim': 'nat'}, {'int': '1'}]}, {'prim': 'ADD', 'annots': ['@valid']}]]}]]}]], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}, {'prim': 'SWAP'}]]}]]}, [[{'prim': 'COMPARE'}, {'prim': 'LE'}], {'prim': 'IF', 'args': [[], [[{'prim': 'UNIT'}, {'prim': 'FAILWITH'}]]]}], {'prim': 'DROP'}, {'prim': 'DROP'}, {'prim': 'DIP', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'PUSH', 'args': [{'prim': 'nat'}, {'int': '1'}]}, {'prim': 'ADD', 'annots': ['@new_counter']}, {'prim': 'PAIR'}]]}, {'prim': 'NIL', 'args': [{'prim': 'operation'}]}, {'prim': 'SWAP'}, {'prim': 'IF_LEFT', 'args': [[[{'prim': 'DUP'}, {'prim': 'CAR'}, {'prim': 'DIP', 'args': [[{'prim': 'CDR'}]]}], {'prim': 'UNIT'}, {'prim': 'TRANSFER_TOKENS'}, {'prim': 'CONS'}], [{'prim': 'IF_LEFT', 'args': [[{'prim': 'SET_DELEGATE'}, {'prim': 'CONS'}], [{'prim': 'DIP', 'args': [[{'prim': 'SWAP'}, {'prim': 'CAR'}]]}, {'prim': 'SWAP'}, {'prim': 'PAIR'}, {'prim': 'SWAP'}]]}]]}, {'prim': 'PAIR'}]]}]`
)

// Entrypoints -
const (
	DefaultEntrypoint  = "default"
	TransferEntrypoint = "transfer"
)

// custom prim
const (
	PrimArray = "_array"
)

// annotations prefix
const (
	AnnotPrefixFieldName     = '%'
	AnnotPrefixrefixTypeName = ':'
)

// indent for printing
const (
	DefaultIndent = "  "
)
