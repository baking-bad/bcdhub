package contract

// Repository -
type Repository interface {
	Get(address string) (Contract, error)
	GetAll(filters map[string]interface{}) ([]Contract, error)
	GetRandom() (Contract, error)
	GetTokens(tokenInterface string, offset, size int64) ([]Contract, int64, error)
	RecentlyCalled(offset, size int64) ([]Contract, error)

	SameCount(c Contract) (int, error)

	Script(address string, symLink string) (Script, error)

	// ScriptPart - returns part of script type. Part can be `storage`, `parameter` or `code`.
	ScriptPart(address string, symLink, part string) ([]byte, error)
}

// ScriptRepository -
type ScriptRepository interface {
	GetScripts(limit, offset int) ([]Script, error)
	ByHash(hash string) (Script, error)
	UpdateProjectID(script []Script) error

	Code(id int64) ([]byte, error)
	Parameter(id int64) ([]byte, error)
	Storage(id int64) ([]byte, error)
	Views(id int64) ([]byte, error)
}

// ConstantRepository -
type ConstantRepository interface {
	Get(address string) (GlobalConstant, error)
	All(addresses ...string) ([]GlobalConstant, error)
	List(size, offset int64) ([]GlobalConstant, error)
}
