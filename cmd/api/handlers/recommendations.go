package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
)

// Recommendations -
func (ctx *Context) Recommendations(c *gin.Context) {
	subscriptions, err := ctx.DB.ListSubscriptions(ctx.OAUTH.UserID)
	if handleError(c, err, 0) {
		return
	}

	if len(subscriptions) == 0 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	ids := make([]string, 0)
	for i := range subscriptions {
		if subscriptions[i].EntityType == "contract" {
			ids = append(ids, subscriptions[i].EntityID)
		}
	}
	contracts, err := ctx.ES.GetContractsByID(ids)
	if handleError(c, err, 0) {
		return
	}

	tags, prefferedLanguage := getContractsTags(contracts)
	addresses := make([]string, len(contracts))
	for i := range contracts {
		addresses[i] = contracts[i].Address
	}

	recommended, err := ctx.ES.Recommendations(tags, prefferedLanguage, addresses, 5)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, recommended)
}

func getContractsTags(contracts []models.Contract) ([]string, string) {
	languages := map[string]int{}
	tagsMap := map[string]struct{}{}
	for i := range contracts {
		if l, ok := languages[contracts[i].Language]; !ok {
			languages[contracts[i].Language] = 1
		} else {
			l++
			languages[contracts[i].Language] = l
		}

		for _, tag := range contracts[i].Tags {
			if _, ok := tagsMap[tag]; !ok {
				tagsMap[tag] = struct{}{}
			}
		}
	}

	tags := make([]string, 0)
	for tag := range tagsMap {
		tags = append(tags, tag)
	}

	max := 0
	lang := ""
	for l, count := range languages {
		if count > max {
			lang = l
			max = count
		}
	}

	return tags, lang
}
