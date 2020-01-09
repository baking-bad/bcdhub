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
)
