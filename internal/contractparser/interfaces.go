package contractparser

var fa12 = []Entrypoint{
	Entrypoint{
		Name: "transfer",
		Type: "",
		Args: []string{
			"address",
			"address",
			"nat",
		},
	},
	Entrypoint{
		Name: "approve",
		Type: "",
		Args: []string{
			"address",
			"nat",
		},
	},
	Entrypoint{
		Name: "getAllowance",
		Type: "",
		Args: []string{
			"address",
			"address",
			"nat",
		},
	},
	Entrypoint{
		Name: "getBalance",
		Type: "",
		Args: []string{
			"address",
			"nat",
		},
	}, Entrypoint{
		Name: "getTotalSupply",
		Type: "",
		Args: []string{
			"unit",
			"nat",
		},
	},
}
