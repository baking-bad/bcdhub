package macros

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/valyala/fastjson"
)

func getPrim(tree *fastjson.Value) string {
	if tree == nil {
		return ""
	}
	if tree.Type() != fastjson.TypeObject {
		return ""
	}
	if !tree.Exists("prim") {
		return ""
	}
	return tree.Get("prim").String()
}

func getArgs(tree *fastjson.Value) []*fastjson.Value {
	if tree == nil {
		return nil
	}
	if tree.Type() != fastjson.TypeObject {
		return nil
	}
	if !tree.Exists("args") {
		return nil
	}
	return tree.GetArray("args")
}

func isEq(text string) bool {
	return helpers.StringInArray(text, []string{
		pEQ, pNEQ, pLT, pLE, pGT, pGE,
	})
}

func isCmpEq(text string) bool {
	return helpers.StringInArray(text, []string{
		pCMPEQ, pCMPNEQ, pCMPLT, pCMPLE, pCMPGT, pCMPGE,
	})
}
