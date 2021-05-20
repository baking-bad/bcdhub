package database

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
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
	Network           types.Network  `json:"network"`
	OperationHash     string         `json:"operation_hash"`
	Sources           pq.StringArray `gorm:"type:varchar(128)[]" json:"sources"`
}

// ListDeployments -
func (d *db) ListDeployments(userID, limit, offset uint) ([]Deployment, error) {
	var deployments []Deployment

	req := d.Scopes(
		userIDScope(userID),
		pagination(limit, offset),
		createdAtDesc,
	)

	return deployments, req.Find(&deployments).Error
}

// CreateDeployment -
func (d *db) CreateDeployment(dt *Deployment) error {
	return d.Create(dt).Error
}

// GetDeploymentBy -
func (d *db) GetDeploymentBy(opHash string) (*Deployment, error) {
	dt := new(Deployment)
	return dt, d.Raw("SELECT * FROM deployments WHERE operation_hash = ?", opHash).Scan(dt).Error
}

// GetDeploymentsByAddressNetwork -
func (d *db) GetDeploymentsByAddressNetwork(address string, network types.Network) ([]Deployment, error) {
	var deployments []Deployment

	req := d.Scopes(
		addressScope(address),
		networkScope(network),
	)

	return deployments, req.Find(&deployments).Error
}

// UpdateDeployment -
func (d *db) UpdateDeployment(dt *Deployment) error {
	return d.Save(dt).Error
}

// CountDeployments -
func (d *db) CountDeployments(userID uint) (int64, error) {
	var count int64
	return count, d.Model(&Deployment{}).Scopes(userIDScope(userID)).Count(&count).Error
}
