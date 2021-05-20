package database

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/jinzhu/gorm"
)

// Assessments -
type Assessments struct {
	Address1   string        `gorm:"primary_key;not null" json:"address_1"`
	Network1   types.Network `gorm:"primary_key;not null" json:"network_1"`
	Address2   string        `gorm:"primary_key;not null" json:"address_2"`
	Network2   types.Network `gorm:"primary_key;not null" json:"network_2"`
	UserID     uint          `gorm:"primary_key;not_null;auto_increment:false" json:"-"`
	Assessment uint          `gorm:"not null" json:"-"`
}

// Assessment field values
const (
	AssessmentUndefined  = uint(10)
	AssessmentSimilar    = uint(1)
	AssessmentNotSimilar = uint(2)
)

// CreateAssessment -
func (d *db) CreateAssessment(a *Assessments) error {
	return d.
		Attrs(Assessments{Assessment: a.Assessment}).
		FirstOrCreate(a).Error
}

// CreateOrUpdateAssessment -
func (d *db) CreateOrUpdateAssessment(a *Assessments) error {
	return d.
		Assign(Assessments{Assessment: a.Assessment}).
		FirstOrCreate(a).Error
}

// GetAssessmentsWithValue -
func (d *db) GetAssessmentsWithValue(userID, assessment, size uint) (result []Assessments, err error) {
	a := &Assessments{
		UserID: userID,
	}
	if assessment == AssessmentUndefined || assessment == AssessmentSimilar || assessment == AssessmentNotSimilar {
		a.Assessment = assessment
	}
	query := d.
		Where(a).
		Order(gorm.Expr("random()"))

	if size > 0 {
		query = query.Limit(size)
	}
	err = query.Find(&result).Error
	return
}

// GetUserCompletedAssesments -
func (d *db) GetUserCompletedAssesments(userID uint) (count int, err error) {
	err = d.Model(&Assessments{}).
		Where("user_id = ? AND (assessment = ? OR assessment = ?)", userID, AssessmentSimilar, AssessmentNotSimilar).
		Count(&count).Error
	return
}
