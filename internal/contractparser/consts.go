package contractparser

// Languages
const (
	LangPython    = "python"
	LangLiquidity = "liquidity"
	LangLigo      = "ligo"
	LangUnknown   = "michelson"
)

// Tags name
const (
	ViewMethodTag      = "view_method"
	ContractFactoryTag = "contract_factory"
	DelegatableTag     = "delegatable"
	ChainAwareTag      = "chain_aware"
	CheckSigTag        = "checksig"
	FA12Tag            = "fa12"
	TestingTag         = "testing"
	VestingTag         = "vesting"
	SpendableTag       = "spendable"
)

const (
	keyPrim   = "prim"
	keyArgs   = "args"
	keyAnnots = "annots"
	keyString = "string"
	keyMutez  = "mutez"
	keyBytes  = "bytes"
	keyInt    = "int"
	keyTime   = "timestamp"
)

// Primitives
const (
	OR             = "OR"
	PAIR           = "PAIR"
	LAMBDA         = "LAMBDA"
	CONTRACT       = "CONTRACT"
	OPTION         = "OPTION"
	ADDRESS        = "ADDReSS"
	NAT            = "NAT"
	STRING         = "STRING"
	UNIT           = "UNIT"
	PARAMETER      = "PARAMETER"
	STORAGE        = "STORAGE"
	CODE           = "CODE"
	CREATECONTRACT = "CREATE_CONTRACT"
	SETDELEGATE    = "SET_DELEGATE"
	CHECKSIGNATURE = "CHECK_SIGNATURE"
	CHAINID        = "CHAIN_ID"
	MAP            = "MAP"
	BIGMAP         = "BIG_MAP"
	SOME           = "SOME"
	LEFT           = "LEFT"
	RIGHT          = "RIGHT"
	IF             = "IF"
	SET            = "SET"
	LIST           = "LIST"
	KEY            = "KEY"
	SIGNATURE      = "SIGNATURE"
	OPERATION      = "OPERATION"
	INT            = "INT"
	BYTES          = "BYTES"
	MUTEZ          = "MUTEZ"
	BOOL           = "BOOL"
	KEYHASH        = "KEY_HASH"
	TIMESTAMP      = "TIMESTAMP"
	PUSH           = "PUSH"
)

// Metadata network names
const (
	MetadataAlpha   = "alpha"
	MetadataBabylon = "babylon"
)

// Babylon
const (
	LevelBabylon = 655360
	HashBabylon  = "PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS"
)
