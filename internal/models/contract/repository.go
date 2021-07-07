package contract

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, address string) (Contract, error)
	GetMany(by map[string]interface{}) ([]Contract, error)
	GetRandom(network types.Network) (Contract, error)
	GetTokens(network types.Network, tokenInterface string, offset, size int64) ([]Contract, int64, error)
	GetProjectsLastContract(c Contract, size, offset int64) ([]*Contract, error)
	GetSameContracts(contact Contract, manager string, size, offset int64) (SameResponse, error)
	GetSimilarContracts(Contract, int64, int64) ([]Similar, int, error)
	Stats(c Contract) (Stats, error)
	GetProjectIDByHash(hash string) (result string, err error)
}
