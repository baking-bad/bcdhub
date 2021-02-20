package interfaces

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// ViewBalanceOf -
type ViewBalanceOf struct{}

// GetName -
func (f *ViewBalanceOf) GetName() string {
	return consts.ViewBalanceOfTag
}

// GetContractInterface -
func (f *ViewBalanceOf) GetContractInterface() string {
	return `{
		"entrypoints": {
			"default": {
				"prim": "list",
				"args": [
					{
						"prim": "pair",
						"args": [
							{
								"prim": "pair",
								"args": [
									{
										"prim": "address"
									},
									{
										"prim": "nat"
									}
								]
							},
							{
								"prim": "nat"
							}
						]
					}
				]
			}
		},
		"is_root": true
	}`
}
