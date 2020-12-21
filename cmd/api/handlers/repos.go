package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/providers"
	"github.com/gin-gonic/gin"
)

// ListPublicAccounts -
func (ctx *Context) ListPublicAccounts(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	orgs, err := getPublicOrgs(user)
	if ctx.handleError(c, err, 0) {
		return
	}

	accounts := []providers.Account{
		{
			Login:     user.Login,
			AvatarURL: user.AvatarURL,
		},
	}

	c.JSON(http.StatusOK, append(accounts, orgs...))
}

func getPublicOrgs(user *database.User) ([]providers.Account, error) {
	if user.Provider == "" {
		return nil, fmt.Errorf("getPublicOrgs error, user has no provider")
	}

	provider, err := providers.NewPublic(user.Provider)
	if err != nil {
		return nil, err
	}

	if user.Login == "" {
		return nil, fmt.Errorf("getPublicOrgs error, user has empty login")
	}

	return provider.GetOrganizations(user.Login)
}

// ListPublicRepos -
func (ctx *Context) ListPublicRepos(c *gin.Context) {
	var req publicReposRequest
	if err := c.BindQuery(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	repos, err := getPublicRepos(req.Login, user)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, repos)
}

func getPublicRepos(account string, user *database.User) ([]providers.Project, error) {
	if user.Provider == "" {
		return nil, fmt.Errorf("getPublicRepos error, user has no provider")
	}

	provider, err := providers.NewPublic(user.Provider)
	if err != nil {
		return nil, err
	}

	if user.Login == "" {
		return nil, fmt.Errorf("getPublicRepos error, user has empty login")
	}

	return provider.GetRepos(account)
}

// ListPublicRefs -
func (ctx *Context) ListPublicRefs(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if ctx.handleError(c, err, 0) {
		return
	}

	var req publicRefsRequest
	if err := c.BindQuery(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	refs, err := getPublicRefs(user, req.Owner, req.Repo)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, refs)
}

func getPublicRefs(user *database.User, owner, repo string) ([]providers.Ref, error) {
	if user.Provider == "" {
		return nil, fmt.Errorf("getPublicRefs error, user has no provider")
	}

	provider, err := providers.NewPublic(user.Provider)
	if err != nil {
		return nil, err
	}

	if user.Login == "" {
		return nil, fmt.Errorf("getPublicRefs error, user has empty login")
	}

	return provider.GetRefs(owner, repo)
}
