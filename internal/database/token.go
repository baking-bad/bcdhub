package database

// Token -
type Token struct {
	ID       uint   `gorm:"primary_key" json:"-"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals uint   `json:"decimals"`
	Contract string `json:"contract"`
	Network  string `json:"network"`
	TokenID  uint   `json:"token_id"`
	DAppID   uint   `json:"-"`
}

// GetTokens -
func (d *db) GetTokens() ([]Token, error) {
	var tokens []Token

	if err := d.ORM.Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}
