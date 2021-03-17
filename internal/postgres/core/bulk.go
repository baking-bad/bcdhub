package core

import (
	"reflect"

	"github.com/baking-bad/bcdhub/internal/models"
	"gorm.io/gorm"
)

// BulkInsert -
func (p *Postgres) BulkInsert(items []models.Model) error {
	return p.DB.Transaction(func(tx *gorm.DB) error {
		for i := range items {
			el := reflect.ValueOf(items[i]).Interface()
			if err := tx.Create(el).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkUpdate -
func (p *Postgres) BulkUpdate(items []models.Model) error {
	return p.DB.Transaction(func(tx *gorm.DB) error {
		for i := range items {
			el := reflect.ValueOf(items[i]).Interface()
			if err := tx.Save(el).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkDelete -
func (p *Postgres) BulkDelete(items []models.Model) error {
	return p.DB.Transaction(func(tx *gorm.DB) error {
		for i := range items {
			el := reflect.ValueOf(items[i]).Interface()
			if err := tx.Delete(el).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
