package core

import (
	"reflect"

	"github.com/baking-bad/bcdhub/internal/models"
)

// UpdateDoc - updates document
func (r *Reindexer) UpdateDoc(model models.Model) error {
	_, err := r.Update(model.GetIndex(), model)
	return err
}

// UpdateFields -
func (r *Reindexer) UpdateFields(index, id string, data interface{}, fields ...string) error {
	query := r.Query(index).Match("id", id)
	for j := range fields {
		value := r.getFieldValue(data, fields[j])
		query = query.Set(fields[j], value)
	}
	it := query.Update()
	defer it.Close()
	return it.Error()
}

func (r *Reindexer) getFieldValue(data interface{}, field string) interface{} {
	val := reflect.ValueOf(data)
	f := reflect.Indirect(val).FieldByName(field)
	return f.Interface()
}
