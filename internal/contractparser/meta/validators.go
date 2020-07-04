package meta

import (
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/btcsuite/btcutil/base58"
)

type validator interface {
	Validate(value interface{}) bool
}

func validate(typ string, value interface{}) bool {
	var valid validator
	switch typ {
	case consts.BYTES:
		valid = &bytesValidator{}
	case consts.ADDRESS:
		valid = &addressValidator{}
	case consts.SIGNATURE:
		valid = &base58Validator{}
	default:
		return true
	}
	return valid.Validate(value)
}

type bytesValidator struct{}

func (v *bytesValidator) Validate(value interface{}) bool {
	sValue, ok := value.(string)
	if !ok {
		return false
	}

	re := regexp.MustCompile(`^([a-f0-9][a-f0-9])*$`)
	return re.MatchString(sValue)
}

type addressValidator struct{}

func (v *addressValidator) Validate(value interface{}) bool {
	address, ok := value.(string)
	if !ok {
		return false
	}
	if !(strings.HasPrefix(address, "KT") || strings.HasPrefix(address, "tz")) || len(address) != 36 {
		return false
	}
	_, _, err := base58.CheckDecode(address)
	return err == nil
}

type base58Validator struct{}

func (v *base58Validator) Validate(value interface{}) bool {
	sValue, ok := value.(string)
	if !ok {
		return false
	}
	_, _, err := base58.CheckDecode(sValue)
	return err == nil
}
