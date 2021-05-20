package core

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
)

// Network -
func Network(network types.Network) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network)
	}
}

// Address -
func Address(address string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("address = ?", address)
	}
}

// Contract -
func Contract(address string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("contract = ?", address)
	}
}

// NetworkAndAddress -
func NetworkAndAddress(network types.Network, address string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network).Where("address = ?", address)
	}
}

// NetworkAndContract -
func NetworkAndContract(network types.Network, address string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network).Where("contract = ?", address)
	}
}

// OrderByLevelDesc -
func OrderByLevelDesc(db *gorm.DB) *gorm.DB {
	return db.Order("level desc")
}

// IsApplied -
func IsApplied(db *gorm.DB) *gorm.DB {
	return db.Where("status = 'applied'")
}

// Token -
func Token(network types.Network, contract string, tokenID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network).
			Where("contract = ?", contract).
			Where("token_id = ?", tokenID)
	}
}
