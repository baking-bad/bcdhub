package database

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/jinzhu/gorm"
)

func addressScope(address string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("address = ?", address)
	}
}

func networkScope(network types.Network) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network)
	}
}

func userIDScope(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

func idScope(id uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

func pagination(limit, offset uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if limit > 0 {
			db = db.Limit(limit)
		}
		if limit > 0 {
			db = db.Offset(offset)
		}
		return db
	}
}

func contract(address string, network types.Network) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("address = ? AND network = ?", address, network)
	}
}

func createdAtDesc(db *gorm.DB) *gorm.DB {
	return db.Order("created_at desc")
}

func loginScope(login string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("login = ?", login)
	}
}
