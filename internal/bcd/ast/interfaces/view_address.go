package interfaces

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// ViewAddress -
type ViewAddress struct{}

// GetName -
func (f *ViewAddress) GetName() string {
	return consts.ViewAddressTag
}

// GetContractInterface -
func (f *ViewAddress) GetContractInterface() string {
	return `{
		"entrypoints": {
			"default": {
				"prim": "address"
			}
		},
		"is_root": true
	}`
}
