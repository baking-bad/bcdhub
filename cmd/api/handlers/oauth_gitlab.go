package handlers

import (
	"fmt"
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

// GitlabOauthLogin -
func (ctx *Context) GitlabOauthLogin(c *gin.Context) {
	url := ctx.OAUTH.Gitlab.AuthCodeURL(ctx.OAUTH.State)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitlabOauthCallback -
func (ctx *Context) GitlabOauthCallback(c *gin.Context) {
	var req OauthRequest
	if err := c.ShouldBind(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.State != ctx.OAUTH.State {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid oauth state"))
		return
	}

	token, err := ctx.OAUTH.Gitlab.Exchange(oauth2.NoContext, req.Code)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("code exchange failed: %s", err.Error()))
		return
	}

	u, _, err := getGitlabUser(token.AccessToken)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("getGithubUser failed: %s", err.Error()))
		return
	}

	user := database.User{
		Token:     token.AccessToken,
		Login:     u.Username,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
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

	location := fmt.Sprintf("http://localhost:14000/v1/oauth/welcome?jwt=%v", jwt)
	c.Redirect(http.StatusTemporaryRedirect, location)
}

func getGitlabUser(token string) (*gitlab.User, *gitlab.Response, error) {
	client := gitlab.NewOAuthClient(nil, token)

	return client.Users.CurrentUser()
}
