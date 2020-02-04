package enrichment

import (
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Alpha -
type Alpha struct {
}

// Do -
func (a Alpha) Do(storage string, bmd gjson.Result) (gjson.Result, error) {
	if bmd.IsArray() && len(bmd.Array()) == 0 {
		return gjson.Parse(storage), nil
	}

	p := miguel.GetGJSONPath("0")

	res := make([]interface{}, 0)
	for _, b := range bmd.Array() {
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 1)
		args[0] = b.Get("key").Value()

		sVal := b.Get("value").String()
		if sVal != "" {
			val := gjson.Parse(sVal)
			args = append(args, val.Value())
		}

		elt["args"] = args
		res = append(res, elt)
	}
	value, err := sjson.Set(storage, p, res)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(value), nil
}

// Level -
func (a Alpha) Level() int64 {
	return 0
}
