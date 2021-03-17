package core

import "gorm.io/gorm"

// Network -
func Network(network string) func(db *gorm.DB) *gorm.DB {
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

// NetworkAndAddress -
func NetworkAndAddress(network, address string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network).Where("address = ?", address)
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
func Token(network, contract string, tokenID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network).
			Where("contract = ?", contract).
			Where("token_id = ?", tokenID)
	}
}
