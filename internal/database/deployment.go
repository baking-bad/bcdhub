package database

import (
	"time"

	"github.com/lib/pq"
)

// Deployment -
type Deployment struct {
	ID                uint           `gorm:"primary_key" json:"id"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         *time.Time     `sql:"index" json:"-"`
	UserID            uint           `json:"user_id"`
	CompilationTaskID uint           `json:"-"`
	Address           string         `json:"address"`
	Network           string         `json:"network"`
	OperationHash     string         `json:"operation_hash"`
	Sources           pq.StringArray `gorm:"type:varchar(128)[]" json:"sources"`
}

// ListDeployments -
func (d *db) ListDeployments(userID, limit, offset uint) ([]Deployment, error) {
	var deployments []Deployment

	req := d.ORM.Where("user_id = ?", userID).Order("created_at desc")

	if limit > 0 {
		req = req.Limit(limit)
	}

	if offset > 0 {
		req = req.Offset(offset)
	}

	if err := req.Find(&deployments).Error; err != nil {
		return nil, err
	}

	return deployments, nil
}

// CreateDeployment -
func (d *db) CreateDeployment(dt *Deployment) error {
	return d.ORM.Create(dt).Error
}

// GetDeploymentBy -
func (d *db) GetDeploymentBy(opHash string) (*Deployment, error) {
	dt := new(Deployment)

	return dt, d.ORM.Where("operation_hash = ?", opHash).First(dt).Error
}

// UpdateDeployment -
func (d *db) UpdateDeployment(dt *Deployment) error {
	return d.ORM.Save(dt).Error
}

// CountDeployments -
func (d *db) CountDeployments(userID uint) (int64, error) {
	var count int64
	return count, d.ORM.Model(&Deployment{}).Where("user_id = ?", userID).Count(&count).Error
}
