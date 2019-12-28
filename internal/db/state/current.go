package state

import (
	"github.com/jinzhu/gorm"
)

// Current -
func Current(d *gorm.DB, network string) (s State, err error) {
	err = d.Where(State{Network: network}).FirstOrCreate(&s).Error
	return
}
