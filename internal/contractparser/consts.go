package contractparser

// Languages
const (
	LangPython    = "python"
	LangLiquidity = "liquidity"
	LangLigo      = "ligo"
	LangUnknown   = "michelson"
)

// Default hashes
const (
	HashVestedContract    = "125176848215215482149ab8bc448125d84484125d841011f41011f413881514164414417101118812e1514191a151416418441011f8e1011f131514191a10111812121b141654b54b441511c191a10111881512212141191d101118481e54b125544bb82b441524101f20881252b4b43112511b1419211011181b14412511b141921101118422191a1011181b1416e191a101118812248b54b4232033b51515421512124192510111851416b821251416bbb82b7651512121214151191d101118411c191a10111881e54b54b5110111248b821211b141689ab54bb444181215281258bbc441710111876445101f2012b3"
	HashDelegatorContract = "1254261427192810111814829a2a1928101118102db1376b3"
	HashTestContract      = "276b"
)

// Kind
const (
	KindTest      = "test"
	KindDelegator = "delegator"
	KindVested    = "vested"
	KindSmart     = "smart"
)

// Tags name
const (
	ViewMethodTag      = "view_method"
	ContractFactoryTag = "contract_factory"
	DelegatableTag     = "delegatable"
	ChainAwareTag      = "chain_aware"
	CheckSigTag        = "checksig"
	FA12Tag            = "fa12"
)
