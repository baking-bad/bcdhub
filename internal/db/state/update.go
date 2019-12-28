package state

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Update -
func Update(d *gorm.DB, level int64, timestamp time.Time) error {
	s := State{
		ID:        1,
		Level:     level,
		Timestamp: timestamp,
	}
	return d.Model(State{}).Updates(s).Error
}
