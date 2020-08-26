package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

// OauthLogin -
func (ctx *Context) OauthLogin(c *gin.Context) {
	var params OauthParams
	if err := c.BindUri(&params); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var redirectURL string

	switch params.Provider {
	case "github":
		redirectURL = ctx.OAUTH.Github.AuthCodeURL(ctx.OAUTH.State)
	case "gitlab":
		redirectURL = ctx.OAUTH.Gitlab.AuthCodeURL(ctx.OAUTH.State)
	default:
		handleError(c, errors.Errorf("invalid provider %v", params.Provider), http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// OauthCallback -
func (ctx *Context) OauthCallback(c *gin.Context) {
	var params OauthParams
	if err := c.BindUri(&params); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var req OauthRequest
	if err := c.ShouldBind(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	if req.State != ctx.OAUTH.State {
		handleError(c, errors.Errorf("invalid oauth state"), http.StatusBadRequest)
		return
	}

	var user database.User
	var err error

	switch params.Provider {
	case "github":
		user, err = ctx.authGithubUser(req.Code)
	case "gitlab":
		user, err = ctx.authGitlabUser(req.Code)
	default:
		handleError(c, errors.Errorf("invalid provider %v", params.Provider), http.StatusBadRequest)
		return
	}

	if handleError(c, err, http.StatusBadRequest) {
		return
	}

	if err := ctx.DB.GetOrCreateUser(&user); handleError(c, err, http.StatusBadRequest) {
		return
	}

	jwt, err := ctx.OAUTH.MakeJWT(user.ID)
	if handleError(c, err, http.StatusBadRequest) {
		return
	}

	location := fmt.Sprintf("%v?jwt=%v", ctx.OAUTH.JWTRedirectURL, jwt)
	c.Redirect(http.StatusTemporaryRedirect, location)
}

func (ctx *Context) authGithubUser(code string) (database.User, error) {
	var user database.User

	token, err := ctx.OAUTH.Github.Exchange(context.Background(), code)
	if err != nil {
		return user, errors.Errorf("github code exchange failed: %s", err.Error())
	}

	u, _, err := getGithubUser(token)
	if err != nil {
		return user, errors.Errorf("getGithubUser failed: %s", err.Error())
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
	}

	return user, nil
}

func getGithubUser(token *oauth2.Token) (*github.User, *github.Response, error) {
	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return client.Users.Get(context.Background(), "")
}

func (ctx *Context) authGitlabUser(code string) (database.User, error) {
	var user database.User

	token, err := ctx.OAUTH.Gitlab.Exchange(context.Background(), code)
	if err != nil {
		return user, errors.Errorf("gitlab code exchange failed: %s", err.Error())
	}

	u, _, err := getGitlabUser(token.AccessToken)
	if err != nil {
		return user, errors.Errorf("getGitlabUser failed: %s", err.Error())
	}

	user = database.User{
		Token:     token.AccessToken,
		Login:     u.Username,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
	}

	return user, nil
}

func getGitlabUser(token string) (*gitlab.User, *gitlab.Response, error) {
	client := gitlab.NewOAuthClient(nil, token)

	return client.Users.CurrentUser()
}
