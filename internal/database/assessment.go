package database

import "github.com/jinzhu/gorm"

// Assessments -
type Assessments struct {
	Address1   string `gorm:"primary_key;not null" json:"address_1"`
	Network1   string `gorm:"primary_key;not null" json:"network_1"`
	Address2   string `gorm:"primary_key;not null" json:"address_2"`
	Network2   string `gorm:"primary_key;not null" json:"network_2"`
	UserID     uint   `gorm:"primary_key;not_null;auto_increment:false" json:"-"`
	Assessment uint   `gorm:"not null" json:"-"`
}

// Assessment field values
const (
	AssessmentUndefined  = 10
	AssessmentSimilar    = 1
	AssessmentNotSimilar = 0
)

// CreateAssessment -
func (d *db) CreateAssessment(a *Assessments) error {
	return d.ORM.Create(a).Error
}

// UpdateAssessment -
func (d *db) CreateOrUpdateAssessment(a *Assessments) error {
	return d.ORM.Where(
		"address1 = ? AND network1 = ? AND address2 = ? AND network2 = ? AND user_id = ?",
		a.Address1, a.Network1, a.Address2, a.Network2, a.UserID).
		Assign(Assessments{Assessment: a.Assessment}).
		FirstOrCreate(a).Error
}

func (d *db) GetNextAssessmentWithValue(userID, assessment uint) (Assessments, error) {
	var a Assessments
	if err := d.ORM.Where("user_id = ? AND assessment = ?", userID, assessment).
		Order(gorm.Expr("random()")).
		First(&a).Error; err != nil {
		return a, err
	}
	return a, nil
}
