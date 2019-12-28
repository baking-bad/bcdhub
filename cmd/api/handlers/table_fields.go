package handlers

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func getTableColumn(db *gorm.DB, table, column, where string, acceptedFields []string) (field interface{}, err error) {
	found := false
	for _, name := range acceptedFields {
		if column == name {
			found = true
			break
		}
	}

	if !found {
		err = fmt.Errorf("Unknown field: %s", column)
		return
	}

	err = db.Select(column).Table(table).Where(where).Row().Scan(&field)
	return
}
