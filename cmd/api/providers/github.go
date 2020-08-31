package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	oauthgithub "golang.org/x/oauth2/github"
)

// Github -
type Github struct {
	Config *oauth2.Config
}

// Name -
func (p *Github) Name() string {
	return "github"
}

// Init -
func (p *Github) Init(cfg config.Config) {
	p.Config = &oauth2.Config{
		RedirectURL:  cfg.OAuth.Github.CallbackURL,
		ClientID:     cfg.OAuth.Github.ClientID,
		ClientSecret: cfg.OAuth.Github.Secret,
		Scopes:       []string{},
		Endpoint:     oauthgithub.Endpoint,
	}
}

// AuthCodeURL -
func (p *Github) AuthCodeURL(state string) string {
	return p.Config.AuthCodeURL(state)
}

// AuthUser -
func (p *Github) AuthUser(code string) (database.User, error) {
	var user database.User

	token, err := p.Config.Exchange(context.Background(), code)
	if err != nil {
		return user, fmt.Errorf("github code exchange failed: %s", err.Error())
	}

	u, _, err := getGithubUser(token)
	if err != nil {
		return user, fmt.Errorf("getGithubUser failed: %s", err.Error())
	}

	var name string
	if u.Name == nil {
		name = *u.Login
	} else {
		name = *u.Name
	}

	user = database.User{
		Token:     token.AccessToken,
		Login:     *u.Login,
		Name:      name,
		AvatarURL: *u.AvatarURL,
		Provider:  p.Name(),
	}

	return user, nil
}

func getGithubUser(token *oauth2.Token) (*github.User, *github.Response, error) {
	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return client.Users.Get(context.Background(), "")
}

// GetRepos -
func (p *Github) GetRepos(login string) ([]Project, error) {
	client := github.NewClient(nil)
	repos, resp, err := client.Repositories.List(context.Background(), login, nil)
	if resp.StatusCode == http.StatusNotFound {
		return []Project{}, nil
	}
	if err != nil {
		return nil, err
	}

	res := make([]Project, len(repos))

	for i, r := range repos {
		res[i] = Project{
			User:    *r.Owner.Login,
			Project: *r.Name,
			URL:     *r.HTMLURL,
		}
	}

	return res, nil
}

// GetRefs -
func (p *Github) GetRefs(owner, repo string) ([]Ref, error) {
	client := github.NewClient(nil)

	branches, err := p.getRefs(client, owner, repo, "heads", "refs/heads/", RefTypeBranch)
	if err != nil {
		return nil, err
	}

	tags, err := p.getRefs(client, owner, repo, "tags", "refs/tags/", RefTypeTag)
	if err != nil {
		return nil, err
	}

	return append(branches, tags...), nil
}

func (p *Github) getRefs(client *github.Client, owner, repo, optType, prefix, refType string) ([]Ref, error) {
	opts := new(github.ReferenceListOptions)
	opts.Type = optType

	refs, resp, err := client.Git.ListRefs(context.Background(), owner, repo, opts)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return []Ref{}, nil
	}

	res := make([]Ref, len(refs))

	for i, r := range refs {
		ref := strings.TrimPrefix(*r.Ref, prefix)
		res[i] = Ref{
			Name: ref,
			URL:  p.ArchivePath(owner, repo, ref),
			Type: refType,
		}
	}

	return res, nil
}

// ArchivePath -
func (p *Github) ArchivePath(owner, repo, ref string) string {
	return fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", owner, repo, ref)
}
