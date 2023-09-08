package contract

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

//go:generate mockgen -source=$GOFILE -destination=../mock/contract/mock.go -package=contract -typed
type Repository interface {
	Get(address string) (Contract, error)
	GetAll(filters map[string]interface{}) ([]Contract, error)
	GetRandom() (Contract, error)
	GetTokens(tokenInterface string, offset, size int64) ([]Contract, int64, error)
	RecentlyCalled(offset, size int64) ([]Contract, error)
	Count() (int, error)

	Script(address string, symLink string) (Script, error)

	// ScriptPart - returns part of script type. Part can be `storage`, `parameter` or `code`.
	ScriptPart(address string, symLink, part string) ([]byte, error)
	FindOne(tags types.Tags) (Contract, error)
}

//go:generate mockgen -source=$GOFILE -destination=../mock/contract/mock.go -package=contract -typed
type ScriptRepository interface {
	GetScripts(limit, offset int) ([]Script, error)
	ByHash(hash string) (Script, error)
	UpdateProjectID(script []Script) error

	Code(id int64) ([]byte, error)
	Parameter(id int64) ([]byte, error)
	Storage(id int64) ([]byte, error)
	Views(id int64) ([]byte, error)
}

//go:generate mockgen -source=$GOFILE -destination=../mock/contract/mock.go -package=contract -typed
type ConstantRepository interface {
	Get(address string) (GlobalConstant, error)
	All(addresses ...string) ([]GlobalConstant, error)
	List(size, offset int64, orderBy, sort string) ([]ListGlobalConstantItem, error)
	ForContract(address string, size, offset int64) ([]GlobalConstant, error)
	ContractList(address string, size, offset int64) ([]Contract, error)
}

// ListGlobalConstantItem -
type ListGlobalConstantItem struct {
	Timestamp  time.Time `json:"timestamp" pg:"timestamp"`
	Level      int64     `json:"level" pg:"level"`
	Address    string    `json:"address" pg:"address"`
	LinksCount uint64    `json:"links_count" pg:"links_count"`
}
