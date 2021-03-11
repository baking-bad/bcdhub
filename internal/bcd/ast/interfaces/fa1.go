package interfaces

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// Fa1 -
type Fa1 struct{}

// GetName -
func (f *Fa1) GetName() string {
	return consts.FA1Tag
}

// GetContractInterface -
func (f *Fa1) GetContractInterface() string {
	return `{
		"entrypoints": {
			"getBalance": {
				"prim": "pair",
				"args": [
					{
						"prim": "address"
					},
					{
						"args": [
							{
								"prim": "nat"
							}
						],
						"prim": "contract"
					}
				]
			},
			"getTotalSupply": {
				"prim": "pair",
				"args": [
					{
						"prim": "unit"
					},
					{
						"args": [
							{
								"prim": "nat"
							}
						],
						"prim": "contract"
					}
				]
			},
			"transfer": {
				"prim": "pair",
				"args": [
					{
						"prim": "address"
					},
					{
						"args": [
							{
								"prim": "address"
							},
							{
								"prim": "nat"
							}
						],
						"prim": "pair"
					}
				]
			}
		}
	}`
}
