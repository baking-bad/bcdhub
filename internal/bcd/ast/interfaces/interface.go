package interfaces

import (
	stdJSON "encoding/json"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Contract -
type Contract interface {
	GetName() string
	GetContractInterface() string
}

// ContractInterface -
type ContractInterface struct {
	IsRoot      bool `json:"is_root,omitempty"`
	Entrypoints map[string]stdJSON.RawMessage
}

// GetAll - receives all contract interfaces
func GetAll() (map[string]ContractInterface, error) {
	all := []Contract{
		&ViewAddress{},
		&ViewBalanceOf{},
		&ViewNat{},
		&Fa1{},
		&Fa1_2{},
		&Fa2{},
	}

	res := make(map[string]ContractInterface)
	for i := range all {
		name := all[i].GetName()
		str := all[i].GetContractInterface()
		var ci ContractInterface
		if err := json.UnmarshalFromString(str, &ci); err != nil {
			return nil, err
		}
		res[name] = ci
	}
	return res, nil
}
