package contractparser

import (
	"github.com/tidwall/gjson"

	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
)

func isFramed(n gjson.Result) bool {
	prim := n.Get("prim").String()
	if helpers.StringInArray(prim, []string{
		PAIR, LEFT, RIGHT, SOME, OR, OPTION, BIGMAP, MAP, LIST, SET, CONTRACT, LAMBDA,
	}) {
		return true
	} else if helpers.StringInArray(prim, []string{
		KEY, UNIT, SIGNATURE, OPERATION, INT, NAT, STRING, BYTES, MUTEZ, BOOL, KEYHASH, TIMESTAMP, ADDRESS,
	}) {
		return n.Get(keyAnnots).Exists()
	}
	return false
}

func isComplex(n gjson.Result) bool {
	prim := n.Get("prim").String()
	return prim == LAMBDA || prim[:2] == IF
}

func isInline(n gjson.Result) bool {
	prim := n.Get("prim").String()
	return prim == PUSH
}

func isScript(n gjson.Result) bool {
	if !n.IsArray() {
		return false
	}
	for _, item := range n.Array() {
		prim := item.Get("prim").String()
		if !helpers.StringInArray(prim, []string{
			PARAMETER, STORAGE, CODE,
		}) {
			return false
		}
	}
	return true
}

// MichelineToMichelson -
func MichelineToMichelson(n gjson.Result) string {
	return ""
}
