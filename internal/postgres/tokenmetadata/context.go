package tokenmetadata

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"gorm.io/gorm"
)

func buildGetTokenMetadataContext(db *gorm.DB, query *gorm.DB, ctx ...tokenmetadata.GetContext) {
	if len(ctx) == 0 {
		return
	}

	fullQuery := new(gorm.DB)
	for i := range ctx {
		subQuery := db.Where("network = ?", ctx[i].Network).Where("contract = ?", ctx[i].Contract)

		if ctx[i].TokenID != nil {
			subQuery.Where("token_id = ?", *ctx[i].TokenID)
		}
		if ctx[i].MaxLevel > 0 {
			subQuery.Where(fmt.Sprintf("level <= %d", ctx[i].MaxLevel))
		}
		if ctx[i].MinLevel > 0 {
			subQuery.Where(fmt.Sprintf("level > %d", ctx[i].MinLevel))
		}
		if ctx[i].Creator != "" {
			subQuery.Where("creators <@ ?", ctx[i].Creator)
		}
		if fullQuery.Statement == nil {
			fullQuery = db.Where(subQuery)
		} else {
			fullQuery.Or(subQuery)
		}
	}
	query.Where(fullQuery)
}
