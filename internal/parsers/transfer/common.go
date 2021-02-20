package transfer

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func normalizeParameter(params string) gjson.Result {
	parameter := gjson.Parse(params)
	if parameter.Get("value").Exists() {
		parameter = parameter.Get("value")
	}

	for prim := parameter.Get("prim").String(); prim == "Right" || prim == "Left"; prim = parameter.Get("prim").String() {
		parameter = parameter.Get("args.0")
	}
	return parameter
}

func getParameters(str string) gjson.Result {
	parameters := gjson.Parse(str)
	if !parameters.Get("value").Exists() {
		return parameters
	}
	parameters = parameters.Get("value")
	for end := false; !end; {
		prim := parameters.Get("prim|@lower").String()
		end = prim != consts.LEFT && prim != consts.RIGHT
		if !end {
			parameters = parameters.Get("args.0")
		}
	}
	return parameters
}

func getAddress(data gjson.Result) (string, error) {
	if data.Get("string").Exists() {
		return data.Get("string").String(), nil
	}

	if data.Get("bytes").Exists() {
		return forge.UnforgeAddress(data.Get("bytes").String())
	}
	return "", errors.Errorf("Unknown address data: %s", data.Raw)
}
