package contract

import (
	"github.com/jinzhu/gorm"
)

// Add -
func Add(db *gorm.DB, c *Contract) error {
	return db.Create(c).Error
}
