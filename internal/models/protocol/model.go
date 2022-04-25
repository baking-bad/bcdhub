package protocol

import (
	"github.com/go-pg/pg/v10"
)

// Protocol -
type Protocol struct {
	// nolint
	tableName struct{} `pg:"protocols"`

	ID int64

	Hash       string `pg:",unique:protocol"`
	StartLevel int64  `pg:",use_zero"`
	EndLevel   int64  `pg:",use_zero"`
	SymLink    string
	Alias      string
	ChainID    string
	*Constants
}

// Constants -
type Constants struct {
	CostPerByte                  int64 `pg:",use_zero"`
	HardGasLimitPerOperation     int64 `pg:",use_zero"`
	HardStorageLimitPerOperation int64 `pg:",use_zero"`
	TimeBetweenBlocks            int64 `pg:",use_zero"`
}

// GetID -
func (p *Protocol) GetID() int64 {
	return p.ID
}

// GetIndex -
func (p *Protocol) GetIndex() string {
	return "protocols"
}

// Save -
func (p *Protocol) Save(tx pg.DBI) error {
	_, err := tx.Model(p).
		OnConflict("(hash) DO UPDATE").
		Set("end_level = ?", p.EndLevel).
		Returning("id").Insert()
	return err
}

// ValidateChainID -
func (p *Protocol) ValidateChainID(chainID string) bool {
	if p.ChainID == "" {
		return p.StartLevel == 0
	}
	return p.ChainID == chainID
}
