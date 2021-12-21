package contract

import "github.com/go-pg/pg/v10"

// ContractConstants -
type ContractConstants struct {
	// nolint
	tableName struct{} `pg:"contract_constants"`

	ContractId       int64
	GlobalConstantId int64
}

// GetID -
func (ContractConstants) GetID() int64 {
	return 0
}

// GetIndex -
func (ContractConstants) GetIndex() string {
	return "contracts"
}

// Save -
func (ContractConstants) Save(tx pg.DBI) error {
	return nil
}
