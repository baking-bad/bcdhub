package core

import (
	"reflect"

	"github.com/baking-bad/bcdhub/internal/models"
	"gorm.io/gorm"
)

// Save - perform insert or update items
func (p *Postgres) Save(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	return p.DB.Transaction(func(tx *gorm.DB) error {
		for i := range items {
			if err := items[i].Save(tx); err != nil {
				tx.Rollback()
				return err
			}
		}
		return nil
	})
}

// BulkDelete -
func (p *Postgres) BulkDelete(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	return p.DB.Transaction(func(tx *gorm.DB) error {
		for i := range items {
			el := reflect.ValueOf(items[i]).Interface()
			if err := tx.Delete(el).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		return nil
	})
}
