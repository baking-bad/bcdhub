package interfaces

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// ViewNat -
type ViewNat struct{}

// GetName -
func (f *ViewNat) GetName() string {
	return consts.ViewNatTag
}

// GetContractInterface -
func (f *ViewNat) GetContractInterface() string {
	return `{
		"entrypoints": {
			"default": {
				"prim": "nat"
			}
		},
		"is_root": true
	}`
}
