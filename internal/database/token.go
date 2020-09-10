package database

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

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

	MetadataJSON postgres.Jsonb `json:"-"`
	Metadata     TokenMetadata  `gorm:"-" json:"-"`
}

// BeforeSave -
func (token *Token) BeforeSave(tx *gorm.DB) error {
	return token.MetadataJSON.Scan(token.Metadata)
}

// AfterFind -
func (token *Token) AfterFind(tx *gorm.DB) error {
	b, err := token.MetadataJSON.MarshalJSON()
	if err != nil {
		return err
	}
	if string(b) == "null" {
		return nil
	}
	return json.Unmarshal(b, &token.Metadata)
}

// TokenMetadata -
type TokenMetadata struct {
	Version    string
	License    string
	Authors    []string
	Interfaces []string
	Views      []TokenView
}

// TokenView -
type TokenView struct {
	Name            string
	Description     string
	Pure            bool
	Implementations []TokenViewImplementation
}

//TokenViewImplementation -
type TokenViewImplementation struct {
	MichelsonParameterView MichelsonParameterView
}

// MichelsonParameterView -
type MichelsonParameterView struct {
	Parameter   interface{}
	ReturnType  interface{}
	Code        interface{}
	Entrypoints []string
}

// GetTokens -
func (d *db) GetTokens() ([]Token, error) {
	var tokens []Token

	if err := d.ORM.Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}
