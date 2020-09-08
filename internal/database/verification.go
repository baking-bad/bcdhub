package database

import "time"

// Verification -
type Verification struct {
	ID                uint       `gorm:"primary_key" json:"id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `sql:"index" json:"-"`
	UserID            uint       `json:"user_id"`
	CompilationTaskID uint       `json:"-"`
	Address           string     `json:"address"`
	Network           string     `json:"network"`
	SourcePath        string     `json:"source_path"`
}

// ListVerifications -
func (d *db) ListVerifications(userID, limit, offset uint) ([]Verification, error) {
	var verifications []Verification

	req := d.ORM.Where("user_id = ?", userID).Order("created_at desc")

	if limit > 0 {
		req = req.Limit(limit)
	}

	if offset > 0 {
		req = req.Offset(offset)
	}

	if err := req.Find(&verifications).Error; err != nil {
		return nil, err
	}

	return verifications, nil
}

// GetVerificationBy -
func (d *db) GetVerificationBy(address, network string) (*Verification, error) {
	v := new(Verification)

	return v, d.ORM.Where("address = ? AND network = ?", address, network).First(v).Error
}

// CreateVerification -
func (d *db) CreateVerification(v *Verification) error {
	return d.ORM.Create(v).Error
}

// CountVerifications -
func (d *db) CountVerifications(userID uint) (int64, error) {
	var count int64
	return count, d.ORM.Model(&Verification{}).Where("user_id = ?", userID).Count(&count).Error
}
