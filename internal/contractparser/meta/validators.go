package meta

import "regexp"

type validator interface {
	Validate(value interface{}) bool
}

func validate(typ string, value interface{}) bool {
	var valid validator
	if typ == "bytes" {
		valid = &bytesValidator{}
	} else {
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
