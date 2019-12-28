package project

import (
	"github.com/jinzhu/gorm"
)

// Get - return project by id
func Get(db *gorm.DB, id int64) (p Project, err error) {
	err = db.First(&p, id).Error
	return
}
