package core

import (
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/iancoleman/strcase"
	"gorm.io/gorm"
)

// UpdateDoc -
func (p *Postgres) UpdateDoc(model models.Model) error {
	el := reflect.ValueOf(model).Interface()
	return p.DB.Where("id = ?", model.GetID()).Updates(el).Error
}

// UpdateFields -
func (p *Postgres) UpdateFields(index string, id int64, data interface{}, fields ...string) error {
	updates := GetFieldsForModel(data, fields...)
	return p.DB.Table(index).Where("id = ?", id).Updates(updates).Error
}

// SetAlias -
func (p *Postgres) SetAlias(network types.Network, address, alias string) error {
	return p.DB.Transaction(func(tx *gorm.DB) error {
		for _, field := range []string{"source_alias", "destination_alias", "delegate_alias"} {
			query := tx.Model(&operation.Operation{}).
				Select(field).
				Where("network = ?", network)
			var op operation.Operation
			switch field {
			case "source_alias":
				op.SourceAlias = alias
				query.Where("source = ?", address)
			case "destination_alias":
				op.DestinationAlias = alias
				query.Where("destination = ?", address)
			case "delegate_alias":
				query.Where("delegate = ?", address)
				op.DelegateAlias = alias
			}
			if err := query.Updates(&op).Error; err != nil {
				return err
			}
		}
		for _, field := range []string{"alias", "delegate_alias"} {
			query := tx.Model(&contract.Contract{}).
				Select(field).
				Where("network = ?", network)
			var c contract.Contract
			switch field {
			case "alias":
				c.Alias = alias
				query.Where("address = ?", address)
			case "delegate_alias":
				c.DelegateAlias = alias
				query.Where("delegate = ?", address)
			}
			if err := query.Updates(&c).Error; err != nil {
				return err
			}
		}
		return nil
	})

}

// GetFieldsForModel -
func GetFieldsForModel(data interface{}, fields ...string) map[string]interface{} {
	t := reflect.TypeOf(data)
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	mapFields := make(map[string]struct{})
	for i := range fields {
		mapFields[fields[i]] = struct{}{}
	}

	updateFields := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if _, ok := mapFields[field.Name]; !ok {
			continue
		}
		value := val.Field(i)
		tag := field.Tag.Get("pg")
		var tagName string
		if tag != "" {
			tagName = strings.Split(tag, ",")[0]
		}
		if tagName == "" {
			tagName = strcase.ToSnake(field.Name)
		}
		updateFields[tagName] = value.Interface()
	}

	return updateFields
}
