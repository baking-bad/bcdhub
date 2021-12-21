package core

import (
	"context"
	"reflect"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/go-pg/pg/v10"
)

// Save - perform insert or update items
func (p *Postgres) Save(ctx context.Context, items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	return p.DB.RunInTransaction(ctx, func(tx *pg.Tx) error {
		for i := range items {
			if err := items[i].Save(tx); err != nil {
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

	return p.DB.RunInTransaction(ctx, func(tx *pg.Tx) error {
		for i := range items {
			el := reflect.ValueOf(items[i]).Interface()
			if _, err := tx.Model().Table(items[i].GetIndex()).Delete(el); err != nil {
				return err
			}
		}
		return nil
	})
}
