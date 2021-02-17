package contract

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/tidwall/gjson"
)

func isDelegatorContract(code []byte, storage ast.UntypedAST) bool {
	if len(code) == 0 {
		return false
	}
	if !checkStorageIsDelegator(storage) {
		return false
	}
	return checkCodeIsDelegator(code)
}

func checkStorageIsDelegator(storage ast.UntypedAST) bool {
	if len(storage) != 1 {
		return false
	}

	switch {
	case storage[0].StringValue != nil:
		return IsAddress(*storage[0].StringValue)
	case storage[0].BytesValue != nil:
		_, err := forge.UnforgeAddress(*storage[0].BytesValue)
		return err == nil
	default:
		return false
	}
}

func checkCodeIsDelegator(code []byte) bool {
	return string(code) == `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"unit","annots":["%default"]}]}]},{"prim":"storage","args":[{"prim":"key_hash"}]},{"prim":"code","args":[[[[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]],{"prim":"IF_LEFT","args":[[{"prim":"PUSH","args":[{"prim":"mutez"},{"int":"0"}]},{"prim":"AMOUNT"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],[{"prim":"DIP","args":[[{"prim":"DUP"}]]},{"prim":"SWAP"}],{"prim":"IMPLICIT_ACCOUNT"},{"prim":"ADDRESS"},{"prim":"SENDER"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]]}]`
}

func isMultisigContract(data gjson.Result) bool {
	code := data.Get("code")

	if data.String() == "" || !code.Exists() {
		return false
	}

	return checkCodeIsMultisig(code)
}

func checkCodeIsMultisig(code gjson.Result) bool {
	return code.String() == consts.MultisigScript1 ||
		code.String() == consts.MultisigScript2 ||
		code.String() == consts.MultisigScript3
}
