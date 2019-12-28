package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type searchRequest struct {
	Text string `form:"text"`
}

type searchByNameResult struct {
	Name    string `json:"name"`
	ID      int64  `json:"id"`
	Type    string `json:"type"`
	Network string `json:"network,omitempty"`
}

func searchByName(db *gorm.DB, search string) ([]searchByNameResult, error) {
	var result []searchByNameResult
	like := fmt.Sprintf("%%%s%%", search)
	query1 := db.Select("alias as name, id, 'project' as type, '' as network").Table("projects").Where("alias LIKE ?", like)
	query2 := db.Select("address as name, id, 'contract' as type, network").Table("contracts").Where("address LIKE ?", like)

	if err := db.Raw(`? UNION ALL ? `, query1.QueryExpr(), query2.QueryExpr()).Limit(10).Find(&result).Error; gorm.IsRecordNotFoundError(err) {
		return []searchByNameResult{}, nil
	} else if err != nil {
		return nil, err
	}
	return result, nil
}

// Search -
func (ctx *Context) Search(c *gin.Context) {
	var req searchRequest
	if err := c.BindQuery(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result, err := searchByName(ctx.DB, req.Text)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
