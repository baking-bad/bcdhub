package consts

// Languages
const (
	LangSmartPy   = "smartpy"
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

// Primitives
const (
	OR             = "or"
	PAIR           = "pair"
	LAMBDA         = "lambda"
	CONTRACT       = "contract"
	OPTION         = "option"
	ADDRESS        = "address"
	NAT            = "nat"
	STRING         = "string"
	UNIT           = "unit"
	PARAMETER      = "parameter"
	STORAGE        = "storage"
	CODE           = "code"
	CREATECONTRACT = "create_contract"
	SETDELEGATE    = "set_delegate"
	CHECKSIGNATURE = "check_signature"
	CHAINID        = "chain_id"
	MAP            = "map"
	BIGMAP         = "big_map"
	SOME           = "some"
	LEFT           = "left"
	RIGHT          = "right"
	IF             = "if"
	SET            = "set"
	LIST           = "list"
	KEY            = "key"
	SIGNATURE      = "signature"
	OPERATION      = "operation"
	INT            = "int"
	BYTES          = "bytes"
	MUTEZ          = "mutez"
	BOOL           = "bool"
	KEYHASH        = "key_hash"
	TIMESTAMP      = "timestamp"
	PUSH           = "push"
	ELT            = "elt"
	NONE           = "none"
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

// Types
const (
	TypeTuple      = "tuple"
	TypeNamedTuple = "namedtuple"
	TypeEnum       = "enum"
	TypeNamedEnum  = "namedenum"
	TypeUnion      = "union"
	TypeNamedUnion = "namedunion"
)

// Data instructions
const (
	Pair = "Pair"
	Some = "Some"
	None = "None"
)
