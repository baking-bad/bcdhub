package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/gin-gonic/gin"
)

// OauthLogin -
func (ctx *Context) OauthLogin(c *gin.Context) {
	var params OauthParams
	if err := c.BindUri(&params); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var redirectURL string

	if provider, ok := ctx.OAUTH.Providers[params.Provider]; ok {
		redirectURL = provider.AuthCodeURL(ctx.OAUTH.State)
	} else {
		handleError(c, fmt.Errorf("invalid provider %v", params.Provider), http.StatusBadRequest)
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

	if provider, ok := ctx.OAUTH.Providers[params.Provider]; ok {
		user, err = provider.AuthUser(req.Code)
	} else {
		handleError(c, fmt.Errorf("invalid provider %v", params.Provider), http.StatusBadRequest)
		return
	}

	if handleError(c, err, http.StatusBadRequest) {
		return
	}

	if err := ctx.DB.GetOrCreateUser(&user, user.Token); handleError(c, err, http.StatusBadRequest) {
		return
	}

	jwt, err := ctx.OAUTH.MakeJWT(user.ID)
	if handleError(c, err, http.StatusBadRequest) {
		return
	}

	location := fmt.Sprintf("%v?jwt=%v", ctx.OAUTH.JWTRedirectURL, jwt)
	c.Redirect(http.StatusTemporaryRedirect, location)
}
