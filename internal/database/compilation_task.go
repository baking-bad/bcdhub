package database

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// CompilationTask model
// kind: verification or deployment
type CompilationTask struct {
	ID        uint                    `gorm:"primary_key" json:"id"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	DeletedAt *time.Time              `sql:"index" json:"-"`
	UserID    uint                    `json:"user_id"`
	Address   string                  `json:"address"`
	Network   types.Network           `json:"network"`
	Account   string                  `json:"account"`
	Repo      string                  `json:"repo"`
	Ref       string                  `json:"ref"`
	Kind      string                  `gorm:"not null" json:"kind"`
	Status    string                  `gorm:"not null" json:"status"`
	Results   []CompilationTaskResult `json:"results,omitempty"`
}

// CompilationTaskResult -
type CompilationTaskResult struct {
	ID                uint            `gorm:"primary_key;not null" json:"id"`
	CompilationTaskID uint            `json:"-"`
	Status            string          `json:"status"`
	Language          string          `json:"language,omitempty"`
	Path              string          `json:"path"`
	AWSPath           string          `json:"aws_path"`
	Script            *postgres.Jsonb `json:"script,omitempty"`
	Error             string          `json:"error,omitempty"`
	Schema            interface{}     `gorm:"-" json:"schema,omitempty"`
	Typedef           interface{}     `gorm:"-" json:"typedef,omitempty"`
}

func (d *db) ListCompilationTasks(userID, limit, offset uint, kind string) ([]CompilationTask, error) {
	var tasks []CompilationTask

	req := d.Preload("Results").Scopes(userIDScope(userID), pagination(limit, offset), createdAtDesc)

	if kind != "" {
		req = req.Where("kind = ?", kind)
	}

	if err := req.Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

func (d *db) GetCompilationTask(taskID uint) (*CompilationTask, error) {
	var task CompilationTask

	if err := d.Preload("Results").Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

// GetCompilationTaskBy -
func (d *db) GetCompilationTaskBy(network types.Network, address, status string) (*CompilationTask, error) {
	task := new(CompilationTask)

	return task, d.Preload("Results").Scopes(contract(address, network)).Where("status = ?", status).First(task).Error
}

// CreateCompilationTask -
func (d *db) CreateCompilationTask(ct *CompilationTask) error {
	return d.Create(ct).Error
}

// UpdateTaskStatus -
func (d *db) UpdateTaskStatus(taskID uint, status string) error {
	return d.Model(&CompilationTask{}).Where("id = ?", taskID).Update("status", status).Error
}

// UpdateTaskResults -
func (d *db) UpdateTaskResults(task *CompilationTask, status string, results []CompilationTaskResult) error {
	task.Status = status
	task.Results = results

	return d.Save(task).Error
}

// CountCompilationTasks -
func (d *db) CountCompilationTasks(userID uint) (int64, error) {
	var count int64
	return count, d.Model(&CompilationTask{}).Scopes(userIDScope(userID)).Count(&count).Error
}
