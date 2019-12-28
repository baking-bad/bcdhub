package handlers

import "github.com/jinzhu/gorm"

// Context -
type Context struct {
	DB *gorm.DB
}

// NewContext -
func NewContext(db *gorm.DB) *Context {
	return &Context{
		DB: db,
	}
}
