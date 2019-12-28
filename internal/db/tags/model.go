package tags

// Tag - current indexer state
type Tag struct {
	ID         int64  `gorm:"AUTO_INCREMENT;unique_index;column:id" json:"-"`
	Name       string `gorm:"column:tag" json:"name"`
	ContractID int64  `gorm:"column:contract_id" json:"-"`
}

// TableName - set table name
func (Tag) TableName() string {
	return "tags"
}
