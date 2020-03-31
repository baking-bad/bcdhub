package contractparser

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

// GetContract -
func GetContract(rpc noderpc.Pool, address, network string, level int64, filesDirectory string) (gjson.Result, error) {
	if filesDirectory != "" {
		var postfix string
		if network == consts.Mainnet {
			if level < consts.LevelBabylon && level != 0 {
				postfix = "_alpha"
			} else {
				postfix = "_babylon"
			}
		}

		filePath := fmt.Sprintf("%s/contracts/%s/%s%s.json", filesDirectory, network, address, postfix)
		_, err := os.Stat(filePath)
		if err == nil {
			f, err := os.Open(filePath)
			if err != nil {
				return gjson.Result{}, err
			}
			defer f.Close()

			data, err := ioutil.ReadAll(f)
			if err != nil {
				return gjson.Result{}, err
			}
			return gjson.ParseBytes(data), nil
		}
		if !os.IsNotExist(err) {
			return gjson.Result{}, err
		}
	}
	return rpc.GetContractJSON(address, level)
}

// IsDelegatorContract -
func IsDelegatorContract(data gjson.Result) bool {
	if data.String() == "" {
		return true
	}
	if !data.Get("storage").Exists() || !data.Get("code").Exists() {
		return false
	}
	storage := data.Get("storage")
	if !checkStorageIsDelegator(storage) {
		return false
	}
	code := data.Get("code")
	if !checkCodeIsDelegator(code) {
		return false
	}
	return true
}

func checkStorageIsDelegator(storage gjson.Result) bool {
	if storage.Get("string").Exists() {
		return isAddress(storage.Get("string").String())
	}
	return storage.Get("bytes").Exists()
}

func checkCodeIsDelegator(code gjson.Result) bool {
	return code.String() == `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"unit","annots":["%default"]}]}]},{"prim":"storage","args":[{"prim":"key_hash"}]},{"prim":"code","args":[[[[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]],{"prim":"IF_LEFT","args":[[{"prim":"PUSH","args":[{"prim":"mutez"},{"int":"0"}]},{"prim":"AMOUNT"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],[{"prim":"DIP","args":[[{"prim":"DUP"}]]},{"prim":"SWAP"}],{"prim":"IMPLICIT_ACCOUNT"},{"prim":"ADDRESS"},{"prim":"SENDER"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]]}]`
}
