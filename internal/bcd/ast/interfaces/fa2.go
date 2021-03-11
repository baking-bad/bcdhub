package interfaces

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// Fa2 -
type Fa2 struct{}

// GetName -
func (f *Fa2) GetName() string {
	return consts.FA2Tag
}

// GetContractInterface -
func (f *Fa2) GetContractInterface() string {
	return `{
		"entrypoints": {
			"transfer": {
				"prim": "list",
				"args": [
					{
						"prim": "pair",
						"args": [
							{
								"prim": "address"
							},
							{
								"prim": "list",
								"args": [
									{
										"prim": "pair",
										"args": [
											{
												"prim": "address"
											},
											{
												"prim": "pair",
												"args": [
													{
														"prim": "nat"
													},
													{
														"prim": "nat"
													}
												]
											}
										]
									}
								]
							}
						]
					}
				]
			},
			"balance_of": {
				"prim": "pair",
				"args": [
					{
						"prim": "list",
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
							}
						]
					},
					{
						"prim": "contract",
						"args": [
							{
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
						]
					}
				]
			},
			"update_operators": {
				"prim": "list",
				"args": [
					{
						"prim": "or",
						"args": [
							{
								"prim": "pair",
								"args": [
									{
										"prim": "address"
									},
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
									}
								]
							},
							{
								"prim": "pair",
								"args": [
									{
										"prim": "address"
									},
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
									}
								]
							}
						]
					}
				]
			}
		}
	}`
}
