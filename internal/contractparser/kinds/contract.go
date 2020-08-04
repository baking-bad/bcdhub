package kinds

import (
	"fmt"

	"github.com/tidwall/gjson"
)

// CheckParameterForTags -
func CheckParameterForTags(parameter string) ([]string, error) {
	interfaces, err := Load(ViewAddressName, ViewNatName, ViewBalanceOfName)
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0)
	for tag, kind := range interfaces {
		if len(kind.Entrypoints) != 1 {
			continue
		}
		parsed := gjson.Parse(parameter)
		if ok := compareEntrypointAndParameter(kind.Entrypoints[0], parsed); ok {
			tags = append(tags, tag)
		}
	}
	return tags, nil
}

func compareEntrypointAndParameter(e Entrypoint, parameter gjson.Result) bool {
	if e.Prim != parameter.Get("prim").String() {
		return false
	}
	if e.Name != parameter.Get("name").String() {
		return false
	}
	for i, arg := range e.Args {
		parsedArg := parameter.Get(fmt.Sprintf("args.%d", i))
		if !parsedArg.IsObject() && !parsedArg.IsArray() {
			return false
		}
		if !compareEntrypointAndParameter(arg, parsedArg) {
			return false
		}
	}
	return true
}
