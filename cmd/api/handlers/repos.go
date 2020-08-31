package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/cmd/api/providers"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/gin-gonic/gin"
)

// ListPublicRepos -
func (ctx *Context) ListPublicRepos(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if handleError(c, err, 0) {
		return
	}

	repos, err := getPublicRepos(user)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, repos)
}

func getPublicRepos(user *database.User) ([]providers.Project, error) {
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

	return provider.GetRepos(user.Login)
}

// ListPublicRefs -
func (ctx *Context) ListPublicRefs(c *gin.Context) {
	userID := CurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	user, err := ctx.DB.GetUser(userID)
	if handleError(c, err, 0) {
		return
	}

	var req publicRefsRequest
	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	refs, err := getPublicRefs(user, req.Repo)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, refs)
}

func getPublicRefs(user *database.User, repo string) ([]providers.Ref, error) {
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

	return provider.GetRefs(user.Login, repo)
}
