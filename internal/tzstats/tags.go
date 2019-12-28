package tzstats

import (
	"reflect"
	"strings"
)

const (
	tagName = "tzstats"
)

func getColumns(s interface{}) string {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	columns := make([]string, 0)
	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get(tagName)
		if tag == "" || tag == "-" {
			continue
		}
		columns = append(columns, tag)
	}
	return strings.Join(columns, ",")
}

func getParentType(a interface{}) reflect.Type {
	for t := reflect.TypeOf(a); ; {
		switch t.Kind() {
		case reflect.Ptr, reflect.Slice:
			t = t.Elem()
		default:
			return t
		}
	}
}
