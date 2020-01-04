package contractparser

var fa12 = Entrypoints{
	"transfer": []string{
		"address",
		"address",
		"nat",
	},
	"approve": []string{
		"address",
		"nat",
	},
	"getAllowance": []string{
		"address",
		"address",
		"nat",
	},
	"getBalance": []string{
		"address",
		"nat",
	},
	"getTotalSupply": []string{
		"unit",
		"nat",
	},
}
