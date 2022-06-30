package services

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/go-pg/pg/v10"
)

func save[M models.Constraint](ctx context.Context, db *pg.DB, items []M) error {
	if len(items) == 0 {
		return nil
	}

	return db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		for i := range items {
			if err := items[i].Save(tx); err != nil {
				return err
			}
		}
		return nil
	})
}
