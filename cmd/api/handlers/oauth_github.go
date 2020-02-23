package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// OauthRequest -
type OauthRequest struct {
	State string `form:"state"`
	Code  string `form:"code"`
}

// GithubOauthLogin -
func (ctx *Context) GithubOauthLogin(c *gin.Context) {
	url := ctx.OAUTH.Github.AuthCodeURL(ctx.OAUTH.State)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GithubOauthCallback -
func (ctx *Context) GithubOauthCallback(c *gin.Context) {
	var req OauthRequest
	if err := c.ShouldBind(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.State != ctx.OAUTH.State {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid oauth state"))
		return
	}

	token, err := ctx.OAUTH.Github.Exchange(oauth2.NoContext, req.Code)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("code exchange failed: %s", err.Error()))
		return
	}

	u, _, err := getGithubUser(token)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("getGithubUser failed: %s", err.Error()))
		return
	}

	user := database.User{
		Token:     token.AccessToken,
		Login:     *u.Login,
		Name:      *u.Name,
		AvatarURL: *u.AvatarURL,
	}

	if err := ctx.DB.GetOrCreateUser(&user); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	jwt, err := ctx.OAUTH.MakeJWT(user.ID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	location := fmt.Sprintf("http://localhost:8080/welcome?jwt=%v", jwt)
	c.Redirect(http.StatusTemporaryRedirect, location)
}

func getGithubUser(token *oauth2.Token) (*github.User, *github.Response, error) {
	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return client.Users.Get(context.Background(), "")
}
