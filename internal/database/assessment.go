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

// CreateAssessment -
func (d *db) CreateOrUpdateAssessment(address1, network1, address2, network2 string, userID, assessment uint) error {
	a := Assessments{
		Address1:   address1,
		Network1:   network1,
		Address2:   address2,
		Network2:   network2,
		UserID:     userID,
		Assessment: assessment,
	}
	return d.ORM.Assign(a).FirstOrCreate(&a).Error
}

func (d *db) GetNextAssessmentWithValue(userID, assessment uint) (Assessments, error) {
	var a Assessments
	if err := d.ORM.Where("user_id = ? AND assessment = ?", userID, assessment).Order(gorm.Expr("random()")).First(&a).Error; err != nil {
		return a, err
	}
	return a, nil
}
