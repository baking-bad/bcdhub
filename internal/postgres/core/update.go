package core

import (
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/iancoleman/strcase"
)

// UpdateDoc -
func (p *Postgres) UpdateDoc(model models.Model) error {
	el := reflect.ValueOf(model).Interface()
	return p.DB.Where("id = ?", model.GetID()).Updates(el).Error
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
