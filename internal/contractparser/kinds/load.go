package kinds

import (
	"encoding/json"
)

// Load -
func Load() (map[string][]Entrypoint, error) {
	interfaces := make(map[string][]Entrypoint)
	for _, i := range []IContractKind{
		Fa1{},
		Fa1_2{},
		Fa2{},
	} {
		var e []Entrypoint
		if err := json.Unmarshal([]byte(i.GetJSON()), &e); err != nil {
			return nil, err
		}
		interfaces[i.GetName()] = e
	}
	return interfaces, nil
}
