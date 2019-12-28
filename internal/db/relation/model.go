package relation

import "github.com/aopoltorzhicky/bcdhub/internal/db/account"

// Relation types
const (
	Hardcoded string = "hardcoded"
)

// Relation -
type Relation struct {
	ID         int64  `gorm:"AUTO_INCREMENT;unique_index;column:id" json:"-"`
	AccountID  int64  `gorm:"account_id" josn:"-"`
	RelationID int64  `gorm:"column:relation_id" json:"-"`
	Type       string `gorm:"column:type" json:"type"`

	Relation account.Account `gorm:"foreignkey:ID;association_foreignkey:RelationID" json:"relation"`
}

// TableName - set table name
func (r *Relation) TableName() string {
	return "relations"
}
