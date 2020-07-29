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

// ArchivePath -
func (p *Gitlab) ArchivePath(owner, repo string) string {
	return fmt.Sprintf("https://gitlab.com/%s/%s/-/archive/master/%s-master.zip", owner, repo, repo)
}
