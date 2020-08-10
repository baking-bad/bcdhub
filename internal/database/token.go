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
