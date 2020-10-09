package providers

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/database"
)

// Public -
type Public interface {
	GetOrganizations(login string) ([]Account, error)
	GetRepos(login string) ([]Project, error)
	GetRefs(owner, repo string) ([]Ref, error)
	ArchivePath(owner, repo, ref string) string
	BaseFilePath(owner, repo, ref string) string
}

// Oauth -
type Oauth interface {
	Name() string
	Init(cfg config.Config)
	AuthCodeURL(state string) string
	AuthUser(code string) (database.User, error)
}

// InitOauth -
func InitOauth(cfg config.Config) map[string]Oauth {
	providers := make(map[string]Oauth)

	for _, provider := range []Oauth{new(Github), new(Gitlab)} {
		provider.Init(cfg)
		providers[provider.Name()] = provider
	}

	return providers
}

// NewPublic -
func NewPublic(name string) (Public, error) {
	switch name {
	case "github":
		return new(Github), nil
	case "gitlab":
		return new(Gitlab), nil
	default:
		return nil, fmt.Errorf("unknown provider %s", name)
	}
}
