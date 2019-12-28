package account

// Account -
type Account struct {
	ID      int64  `gorm:"AUTO_INCREMENT;unique_index;column:id" json:"-"`
	Address string `gorm:"address;primary_key" json:"address"`
	Alias   string `gorm:"alias" json:"alias,omitempty"`
	Network string `gorm:"network;primary_key" json:"-"`
}

// TableName - set table name
func (a *Account) TableName() string {
	return "accounts"
}
