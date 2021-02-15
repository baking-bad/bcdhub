package kinds

import (
	jsoniter "github.com/json-iterator/go"

	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// ContractKind -
type ContractKind struct {
	Entrypoints []Entrypoint
	IsRoot      bool
}

// Load - load interfaces by name. If `names` is empty loads all interfaces.
func Load(names ...string) (map[string]ContractKind, error) {
	interfaces := make(map[string]ContractKind)

	items := map[string]IContractKind{
		FA1Name:           Fa1{},
		FA1_2Name:         Fa1_2{},
		FA2Name:           Fa2{},
		ViewNatName:       ViewNat{},
		ViewAddressName:   ViewAddress{},
		ViewBalanceOfName: ViewBalanceOf{},
	}

	if len(names) == 0 {
		for k := range items {
			names = append(names, k)
		}
	}

	for _, name := range names {
		i, ok := items[name]
		if !ok {
			return nil, errors.Errorf("Invalid interface name: %s", name)
		}
		var e []Entrypoint
		if err := json.UnmarshalFromString(i.GetJSON(), &e); err != nil {
			return nil, err
		}
		interfaces[i.GetName()] = ContractKind{
			Entrypoints: e,
			IsRoot:      i.IsRoot(),
		}
	}

	return interfaces, nil
}
