package macros

import (
	"regexp"

	"github.com/tidwall/gjson"
)

var allMacros = []macros{
	newCompareIfMacros(),
	newCompareMacros(),
	newFailMacros(),
	newAssertNoneMacros(),
}

type macros interface {
	Is(where string) bool
	Collapse(data gjson.Result) map[string]interface{}
	GetRegular() string
}

type defaultMacros struct {
	Reg  string
	Name string
}

func (m defaultMacros) Is(where string) bool {
	re := regexp.MustCompile(m.Reg)
	return re.MatchString(where)
}

func (m defaultMacros) GetRegular() string {
	return m.Reg
}

func (m defaultMacros) Collapse(data gjson.Result) map[string]interface{} {
	return data.Value().(map[string]interface{})
}

func replacePrim(where, reg, target string) string {
	re := regexp.MustCompile(reg)
	return re.ReplaceAllString(where, target)
}
