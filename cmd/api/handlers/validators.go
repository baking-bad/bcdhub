package handlers

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
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
	return err == nil
}

// MakeNetworkValidator -
func MakeNetworkValidator(networks []string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		network := fl.Field().String()
		return helpers.StringInArray(network, networks)
	}
}

// OPGValidator -
var OPGValidator validator.Func = func(fl validator.FieldLevel) bool {
	hash := fl.Field().String()
	if !strings.HasPrefix(hash, "o") && len(hash) != 51 {
		return false
	}
	_, _, err := base58.CheckDecode(hash)
	return err == nil
}

// StatusValidator -
var StatusValidator validator.Func = func(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	data := strings.Split(status, ",")
	for i := range data {
		if !helpers.StringInArray(data[i], []string{
			consts.Applied,
			consts.Backtracked,
			consts.Failed,
			consts.Skipped,
		}) {
			return false
		}
	}
	return true
}

// FAVersionValidator -
var FAVersionValidator validator.Func = func(fl validator.FieldLevel) bool {
	version := fl.Field().String()
	return helpers.StringInArray(version, []string{
		consts.FA1Tag,
		consts.FA12Tag,
		consts.FA2Tag,
	})
}
