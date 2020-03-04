package handlers

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/btcsuite/btcutil/base58"
	"gopkg.in/go-playground/validator.v9"
)

// AddressValidator -
var AddressValidator validator.Func = func(fl validator.FieldLevel) bool {
	address := fl.Field().String()
	if !strings.HasPrefix(address, "KT") && !strings.HasPrefix(address, "tz") && len(address) != 36 {
		return false
	}
	_, _, err := base58.CheckDecode(address)
	if err != nil {
		return false
	}
	return true
}

// NetworkValidator -
var NetworkValidator validator.Func = func(fl validator.FieldLevel) bool {
	network := fl.Field().String()
	if helpers.StringInArray(network, []string{
		consts.Mainnet,
		consts.Babylon,
		consts.Carthage,
		consts.Zeronet,
	}) {
		return true
	}
	return false
}

// OPGValidator -
var OPGValidator validator.Func = func(fl validator.FieldLevel) bool {
	hash := fl.Field().String()
	if !strings.HasPrefix(hash, "o") && len(hash) != 51 {
		return false
	}
	_, _, err := base58.CheckDecode(hash)
	if err != nil {
		return false
	}
	return true
}
