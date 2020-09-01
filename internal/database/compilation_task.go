package database

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// CompilationTask model
// kind: Verify, Deploy, Compile
type CompilationTask struct {
	ID        uint                    `gorm:"primary_key" json:"id"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	DeletedAt *time.Time              `sql:"index" json:"-"`
	UserID    uint                    `json:"user_id"`
	Address   string                  `json:"address"`
	Network   string                  `json:"network"`
	SourceURL string                  `json:"source_url"`
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
	Script            *postgres.Jsonb `json:"script,omitempty"`
	Error             string          `json:"error,omitempty"`
}

func (d *db) ListCompilationTasks(userID, limit, offset uint, kind string) ([]CompilationTask, error) {
	var tasks []CompilationTask

	req := d.ORM.Preload("Results").Where("user_id = ?", userID).Order("created_at desc")

	if kind != "" {
		req = req.Where("kind = ?", kind)
	}

	if limit > 0 {
		req = req.Limit(limit)
	}

	if offset > 0 {
		req = req.Offset(offset)
	}

	if err := req.Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

func (d *db) GetCompilationTask(taskID uint) (*CompilationTask, error) {
	var task CompilationTask

	if err := d.ORM.Preload("Results").Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

// CreateCompilationTask -
func (d *db) CreateCompilationTask(ct *CompilationTask) error {
	return d.ORM.Create(ct).Error
}

// UpdateTaskStatus -
func (d *db) UpdateTaskStatus(taskID uint, status string) error {
	return d.ORM.Model(&CompilationTask{}).Where("id = ?", taskID).Update("status", status).Error
}

// UpdateTaskResults -
func (d *db) UpdateTaskResults(task *CompilationTask, status string, results []CompilationTaskResult) error {
	task.Status = status
	task.Results = results

	return d.ORM.Save(task).Error
}
