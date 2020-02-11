package contractparser

import (
	"github.com/tidwall/gjson"
)

type fail struct {
	With string
}

func parseFail(args gjson.Result) *fail {
	if !args.IsArray() {
		return nil
	}
	if len(args.Array()) < 2 {
		return nil
	}

	failWith := args.Get(`#(prim="FAILWITH")`)
	if !failWith.Exists() {
		return nil
	}

	push := args.Get(`#(prim="PUSH").args.#.string`)
	if !push.Exists() {
		return nil
	}

	return &fail{
		With: push.Get("0").String(),
	}
}
