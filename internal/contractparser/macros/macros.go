package macros

import (
	"regexp"

	"github.com/tidwall/gjson"
)

type macros interface {
	Find(data gjson.Result) bool
	Collapse(data gjson.Result)
	Replace(json, path string) (string, error)
}

type defaultMacros struct {
	Reg  string
	Name string
}

func (m defaultMacros) Find(data gjson.Result) bool {
	return false
}

func (m defaultMacros) Collapse(data gjson.Result) {}

func (m defaultMacros) Replace(json, path string) (string, error) {
	return json, nil
}

func getPrim(item gjson.Result) string {
	return item.Get("prim|@upper").String()
}

func isDip(s string) bool {
	re := regexp.MustCompile("D(I)+P")
	return re.MatchString(s)
}

func isPai(s string) bool {
	re := regexp.MustCompile("^P[PAIR]*$")
	return re.MatchString(s)
}
