package interfaces

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// Fa1_2 -
type Fa1_2 struct{}

// GetName -
func (f *Fa1_2) GetName() string {
	return consts.FA12Tag
}

// GetContractInterface -
func (f *Fa1_2) GetContractInterface() string {
	return `{
		"entrypoints": {
			"approve": {
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
			"getAllowance": {
				"prim": "pair",
				"args": [
					{
						"args": [
							{
								"prim": "address"
							},
							{
								"prim": "address"
							}
						],
						"prim": "pair"
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
