package core

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/uptrace/bun"
)

// Save - perform insert or update items
func (p *Postgres) Save(ctx context.Context, items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	return p.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for i := range items {
			if err := items[i].Save(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkDelete -
func (p *Postgres) BulkDelete(ctx context.Context, items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	return p.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for i := range items {
			if _, err := tx.NewDelete().Table(items[i].GetIndex()).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
