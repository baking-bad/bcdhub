package database

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Verification -
type Verification struct {
	ID                uint          `gorm:"primary_key" json:"id"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	DeletedAt         *time.Time    `sql:"index" json:"-"`
	UserID            uint          `json:"user_id"`
	CompilationTaskID uint          `json:"-"`
	Address           string        `json:"address"`
	Network           types.Network `json:"network"`
	SourcePath        string        `json:"source_path"`
}

// ListVerifications -
func (d *db) ListVerifications(userID, limit, offset uint) ([]Verification, error) {
	var verifications []Verification

	req := d.Scopes(userIDScope(userID), pagination(limit, offset), createdAtDesc)

	if err := req.Find(&verifications).Error; err != nil {
		return nil, err
	}

	return verifications, nil
}

// GetVerificationBy -
func (d *db) GetVerificationBy(address string, network types.Network) (*Verification, error) {
	v := new(Verification)

	return v, d.Scopes(contract(address, network)).First(v).Error
}

// CreateVerification -
func (d *db) CreateVerification(v *Verification) error {
	return d.Create(v).Error
}

// CountVerifications -
func (d *db) CountVerifications(userID uint) (int64, error) {
	var count int64
	return count, d.Model(&Verification{}).Scopes(userIDScope(userID)).Count(&count).Error
}
