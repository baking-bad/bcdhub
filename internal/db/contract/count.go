package contract

import (
	"github.com/jinzhu/gorm"
)

// Count - return contract count
func Count(db *gorm.DB) (c int64, err error) {
	err = db.Model(&Contract{}).Count(&c).Error
	return
}
