package tokenmetadata

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
)

func (storage *Storage) buildGetTokenMetadataContext(query *gorm.DB, ctx ...tokenmetadata.GetContext) {
	if len(ctx) == 0 {
		return
	}

	fullQuery := new(gorm.DB)
	for i := range ctx {
		subQuery := storage.DB.Table(models.DocTokenMetadata)

		if ctx[i].Network != types.Empty {
			subQuery.Where("network = ?", ctx[i].Network)
		}
		if ctx[i].Contract != "" {
			subQuery.Where("contract = ?", ctx[i].Contract)
		}
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
		if ctx[i].Name != "" {
			subQuery.Where("name = ?", ctx[i].Name)
		}
		if fullQuery.Statement == nil {
			fullQuery = storage.DB.Where(subQuery)
		} else {
			fullQuery.Or(subQuery)
		}
	}
	query.Where(fullQuery)
}
