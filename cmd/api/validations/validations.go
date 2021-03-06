package validations

import (
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/compiler/compilation"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/btcsuite/btcutil/base58"
	"gopkg.in/go-playground/validator.v9"
)

// Register -
func Register(v *validator.Validate, networks []string) error {
	if err := v.RegisterValidation("address", addressValidator()); err != nil {
		return err
	}

	if err := v.RegisterValidation("opg", opgValidator()); err != nil {
		return err
	}

	if err := v.RegisterValidation("network", networkValidator(networks)); err != nil {
		return err
	}

	if err := v.RegisterValidation("status", statusValidator()); err != nil {
		return err
	}

	if err := v.RegisterValidation("faversion", faVersionValidator()); err != nil {
		return err
	}

	if err := v.RegisterValidation("fill_type", fillTypeValidator()); err != nil {
		return err
	}

	if err := v.RegisterValidation("compilation_kind", compilationKindValidator()); err != nil {
		return err
	}

	if err := v.RegisterValidation("search", searchStringValidator()); err != nil {
		return err
	}

	if err := v.RegisterValidation("gt_int64_ptr", greatThanInt64PtrValidator()); err != nil {
		return err
	}

	return nil
}

func addressValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		address := fl.Field().String()
		if !strings.HasPrefix(address, "KT") && !strings.HasPrefix(address, "tz") && len(address) != 36 {
			return false
		}
		_, _, err := base58.CheckDecode(address)
		return err == nil
	}
}

func networkValidator(networks []string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		network := fl.Field().String()
		return helpers.StringInArray(network, networks)
	}
}

func opgValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		hash := fl.Field().String()
		if !strings.HasPrefix(hash, "o") && len(hash) != 51 {
			return false
		}
		_, _, err := base58.CheckDecode(hash)
		return err == nil
	}
}

func statusValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
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
}

func faVersionValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		version := fl.Field().String()
		return helpers.StringInArray(version, []string{
			consts.FA1Tag,
			"fa12",
			consts.FA2Tag,
		})
	}
}

func fillTypeValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		fillType := fl.Field().String()
		return helpers.StringInArray(fillType, []string{
			"empty",
			"current",
		})
	}
}

func compilationKindValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		kind := fl.Field().String()
		return helpers.StringInArray(kind, []string{
			compilation.KindVerification,
			compilation.KindDeployment,
		})
	}
}

func searchStringValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) > 2
	}
}

func greatThanInt64PtrValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()
		kind := field.Kind()

		currentField, currentKind, _, ok := fl.GetStructFieldOK2()
		if !ok {
			return false
		}

		switch {
		case kind == reflect.Ptr && currentKind == reflect.Ptr:
			return true
		case kind == reflect.Ptr && currentKind == reflect.Int64:
			return true
		case kind == reflect.Int64 && currentKind == reflect.Ptr:
			return true
		case kind == reflect.Int64 && currentKind == reflect.Int64:
			return field.Int() > currentField.Int()
		default:
			return false
		}
	}
}
