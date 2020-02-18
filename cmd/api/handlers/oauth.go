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

// GetOauthWelcome -
func (ctx *Context) GetOauthWelcome(c *gin.Context) {
	jwt := c.Query("jwt")

	c.JSON(http.StatusOK, gin.H{"message": jwt})
}

// GetOauthLogin -
func (ctx *Context) GetOauthLogin(c *gin.Context) {
	url := ctx.OAUTH.Oauth2.AuthCodeURL(ctx.OAUTH.State)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GetOauthCallback -
func (ctx *Context) GetOauthCallback(c *gin.Context) {
	var req OauthRequest
	if err := c.ShouldBind(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.State != ctx.OAUTH.State {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid oauth state"))
		return
	}

	token, err := ctx.OAUTH.Oauth2.Exchange(oauth2.NoContext, req.Code)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("code exchange failed: %s", err.Error()))
		return
	}

	u, _, err := getGithubUser(token)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("getGithubUser failed: %s", err.Error()))
		return
	}

	user := ctx.DB.GetUserByLogin(*u.Login)

	if user.Login == "" {
		usr := database.User{
			Token:     token.AccessToken,
			Login:     *u.Login,
			Name:      *u.Name,
			AvatarURL: *u.AvatarURL,
		}

		err = ctx.DB.CreateUser(usr)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	jwt, err := ctx.OAUTH.MakeJWT(*u.Login)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	location := fmt.Sprintf("http://localhost:14000/v1/oauth/welcome?jwt=%v", jwt)
	c.Redirect(http.StatusTemporaryRedirect, location)
}

func getGithubUser(token *oauth2.Token) (*github.User, *github.Response, error) {
	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return client.Users.Get(context.Background(), "")
}
