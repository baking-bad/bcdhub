package enrichment

import (
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Babylon -
type Babylon struct {
}

// Do -
func (b Babylon) Do(storage string, bmd gjson.Result) (gjson.Result, error) {
	if bmd.IsArray() && len(bmd.Array()) == 0 {
		return gjson.Parse(storage), nil
	}

	data := gjson.Parse(storage)
	for _, b := range bmd.Array() {
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 1)
		val := gjson.Parse(b.Get("value").String())
		args[0] = b.Get("key").Value()

		if b.Get("value").String() != "" {
			args = append(args, val.Value())
		}

		elt["args"] = args

		p := miguel.GetGJSONPath(b.Get("bin_path").String()[2:])
		value, err := sjson.Set(storage, p, []interface{}{elt})
		if err != nil {
			return gjson.Result{}, err
		}
		data = gjson.Parse(value)
	}

	return data, nil
}

// Level -
func (b Babylon) Level() int64 {
	return consts.LevelBabylon
}
