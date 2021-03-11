package interfaces

import (
	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var all = map[string]Contract{
	consts.ViewAddressTag:   &ViewAddress{},
	consts.ViewBalanceOfTag: &ViewBalanceOf{},
	consts.ViewNatTag:       &ViewNat{},
	consts.FA1Tag:           &Fa1{},
	consts.FA12Tag:          &Fa1_2{},
	consts.FA2Tag:           &Fa2{},
}

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
	res := make(map[string]ContractInterface)
	for _, i := range all {
		name := i.GetName()
		str := i.GetContractInterface()
		var ci ContractInterface
		if err := json.UnmarshalFromString(str, &ci); err != nil {
			return nil, err
		}
		res[name] = ci
	}
	return res, nil
}

// GetMethods - returns list of interface methods
func GetMethods(name string) ([]string, error) {
	i, ok := all[name]
	if !ok {
		return nil, errors.Errorf("Unknwon interface name: %s", name)
	}
	var ci ContractInterface
	if err := json.UnmarshalFromString(i.GetContractInterface(), &ci); err != nil {
		return nil, err
	}
	methods := make([]string, len(ci.Entrypoints))
	for entrypoint := range ci.Entrypoints {
		methods = append(methods, entrypoint)
	}
	return methods, nil
}
