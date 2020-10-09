package providers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	oauthgitlab "golang.org/x/oauth2/gitlab"
)

// Gitlab -
type Gitlab struct {
	Config *oauth2.Config
}

// Name -
func (p *Gitlab) Name() string {
	return "gitlab"
}

// Init -
func (p *Gitlab) Init(cfg config.Config) {
	p.Config = &oauth2.Config{
		RedirectURL:  cfg.OAuth.Gitlab.CallbackURL,
		ClientID:     cfg.OAuth.Gitlab.ClientID,
		ClientSecret: cfg.OAuth.Gitlab.Secret,
		Scopes:       []string{"read_user"},
		Endpoint:     oauthgitlab.Endpoint,
	}
}

// AuthCodeURL -
func (p *Gitlab) AuthCodeURL(state string) string {
	return p.Config.AuthCodeURL(state)
}

// AuthUser -
func (p *Gitlab) AuthUser(code string) (database.User, error) {
	var user database.User

	token, err := p.Config.Exchange(context.Background(), code)
	if err != nil {
		return user, fmt.Errorf("gitlab code exchange failed: %s", err.Error())
	}

	u, _, err := getGitlabUser(token.AccessToken)
	if err != nil {
		return user, fmt.Errorf("getGitlabUser failed: %s", err.Error())
	}

	user = database.User{
		Token:     token.AccessToken,
		Login:     u.Username,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		Provider:  p.Name(),
	}

	return user, nil
}

func getGitlabUser(token string) (*gitlab.User, *gitlab.Response, error) {
	client, err := gitlab.NewOAuthClient(token)
	if err != nil {
		return nil, nil, err
	}

	return client.Users.CurrentUser()
}

// GetOrganizations -
func (p *Gitlab) GetOrganizations(login string) ([]Account, error) {
	return []Account{}, nil
}

// GetRepos -
func (p *Gitlab) GetRepos(login string) ([]Project, error) {
	client, err := gitlab.NewClient("")
	if err != nil {
		return nil, err
	}

	projects, resp, err := client.Projects.ListUserProjects(login, nil)
	if resp.StatusCode == http.StatusNotFound {
		return []Project{}, nil
	}
	if err != nil {
		return nil, err
	}

	res := make([]Project, len(projects))

	for i, r := range projects {
		res[i] = Project{
			User:    r.Owner.Name,
			Project: r.Path,
			URL:     r.WebURL,
		}
	}

	return res, nil
}

// GetRefs -
func (p *Gitlab) GetRefs(owner, repo string) ([]Ref, error) {
	client, err := gitlab.NewClient("")
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%s/%s", owner, repo)

	branches, err := p.getBranches(client, id, owner, repo)
	if err != nil {
		return nil, err
	}

	tags, err := p.getTags(client, id, owner, repo)
	if err != nil {
		return nil, err
	}

	return append(branches, tags...), nil
}

func (p *Gitlab) getTags(client *gitlab.Client, id, owner, repo string) ([]Ref, error) {
	tags, resp, err := client.Tags.ListTags(id, nil)
	if resp.StatusCode == http.StatusNotFound {
		return []Ref{}, nil
	}
	if err != nil {
		return nil, err
	}

	res := make([]Ref, len(tags))

	for i, t := range tags {
		res[i] = Ref{
			Name: t.Name,
			URL:  p.ArchivePath(owner, repo, t.Name),
			Type: RefTypeTag,
		}
	}

	return res, nil
}

func (p *Gitlab) getBranches(client *gitlab.Client, id, owner, repo string) ([]Ref, error) {
	branches, resp, err := client.Branches.ListBranches(id, nil)
	if resp.StatusCode == http.StatusNotFound {
		return []Ref{}, nil
	}
	if err != nil {
		return nil, err
	}

	res := make([]Ref, len(branches))

	for i, t := range branches {
		res[i] = Ref{
			Name: t.Name,
			URL:  p.ArchivePath(owner, repo, t.Name),
			Type: RefTypeBranch,
		}
	}

	return res, nil
}

// ArchivePath -
func (p *Gitlab) ArchivePath(owner, repo, ref string) string {
	return fmt.Sprintf("https://gitlab.com/%s/%s/-/archive/%s/%s-%s.zip", owner, repo, ref, repo, ref)
}

// BaseFilePath -
func (p *Gitlab) BaseFilePath(owner, repo, ref string) string {
	return fmt.Sprintf("https://gitlab.com/%s/%s/-/blob/%s/", owner, repo, ref)
}
