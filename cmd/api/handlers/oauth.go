package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type oauthRequest struct {
	State string `form:"state"`
	Code  string `form:"code"`
}

// GetOauthLogin -
func (ctx *Context) GetOauthLogin(c *gin.Context) {
	// TODO: randomize it and move to .env
	oauthStateString = "pseudo-random"

	url := ctx.OAUTH.AuthCodeURL(oauthStateString)

	log.Println("[url]", url)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GetOauthCallback -
func (ctx *Context) GetOauthCallback(c *gin.Context) {
	oauthStateString = "pseudo-random"

	var req oauthRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.State != oauthStateString {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid oauth state"))
		return
	}

	token, err := ctx.OAUTH.Exchange(oauth2.NoContext, req.Code)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("code exchange failed: %s", err.Error()))
		return
	}

	// 1. check token in db
	// TO-DO

	// 1.1. create user
	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	u, _, err := client.Users.Get(ctx, "")
	log.Println(u)
	if err != nil {
		return "", err
	}
	// TO-DO: save user in db

	// 1.2. get user from db
	// TO-DO

	// 2. generate jwt token
	jwt, err := makeJWT(*u.Login)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// 3. redirect to /welcome?jwt=aaasd
	location := fmt.Sprintf("localhost:8080/welcome?jwt=%v", jwt)

	c.Redirect(http.StatusTemporaryRedirect, location)
}

// Claims -
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func makeJWT(username string) (string, error) {
	expirationTime := time.Now().Add(48 * time.Hour)

	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtKey))
}
