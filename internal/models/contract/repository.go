package contract

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, address string) (Contract, error)
	GetMany(network types.Network) ([]Contract, error)
	GetRandom(networks ...types.Network) (Contract, error)
	GetTokens(network types.Network, tokenInterface string, offset, size int64) ([]Contract, int64, error)
	RecentlyCalled(network types.Network, offset, size int64) ([]Contract, error)

	GetSameContracts(contact Contract, manager string, size, offset int64) (SameResponse, error)
	Stats(c Contract) (int, error)

	Script(network types.Network, address string, symLink string) (Script, error)

	// ScriptPart - returns part of script type. Part can be `storage`, `parameter` or `code`.
	ScriptPart(network types.Network, address string, symLink, part string) ([]byte, error)
}

// ScriptRepository -
type ScriptRepository interface {
	GetScripts(limit, offset int) ([]Script, error)
	ByHash(hash string) (Script, error)
	UpdateProjectID(script []Script) error

	Code(id int64) ([]byte, error)
	Parameter(id int64) ([]byte, error)
	Storage(id int64) ([]byte, error)
}
