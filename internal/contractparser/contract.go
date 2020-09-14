package contractparser

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	contractFormatPath = "%s/contracts/%s/%s_%s.json"
)

// GetContract -
func GetContract(rpc noderpc.INode, address, network, protocol, filesDirectory string, fallbackLevel int64) (gjson.Result, error) {
	if filesDirectory != "" {
		protoSymLink, err := meta.GetProtoSymLink(protocol)
		if err != nil {
			return gjson.Result{}, err
		}

		filePath := fmt.Sprintf(contractFormatPath, filesDirectory, network, address, protoSymLink)
		_, err = os.Stat(filePath)
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
			contract := gjson.ParseBytes(data)
			if contract.Get("script").Exists() {
				contract = contract.Get("script")
			}
			return contract, nil
		}
		if !strings.Contains(err.Error(), "no such file or directory") {
			return gjson.Result{}, err
		}
	}
	contract, err := rpc.GetContractJSON(address, fallbackLevel)
	if err != nil {
		return gjson.Result{}, err
	}
	if contract.Get("script").Exists() {
		contract = contract.Get("script")
	}
	return contract, nil
}

// RemoveContractFromFileSystem -
func RemoveContractFromFileSystem(address, network, protocol, filesDirectory string) error {
	if filesDirectory == "" {
		return errors.Errorf("Invalid filesDirectory: %s", filesDirectory)
	}
	protoSymLink, err := meta.GetProtoSymLink(protocol)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf(contractFormatPath, filesDirectory, network, address, protoSymLink)
	if _, err = os.Stat(filePath); err == nil {
		return os.Remove(filePath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// RemoveAllContracts -
func RemoveAllContracts(network, filesDirectory string) error {
	if filesDirectory == "" {
		return errors.Errorf("Invalid filesDirectory: %s", filesDirectory)
	}

	dirPath := fmt.Sprintf("%s/contracts/%s", filesDirectory, network)
	if _, err := os.Stat(dirPath); err == nil {
		return os.RemoveAll(dirPath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
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

	return checkCodeIsDelegator(code)
}

func checkStorageIsDelegator(storage gjson.Result) bool {
	if storage.Get("string").Exists() {
		return helpers.IsAddress(storage.Get("string").String())
	}
	return storage.Get("bytes").Exists()
}

func checkCodeIsDelegator(code gjson.Result) bool {
	return code.String() == `[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"unit","annots":["%default"]}]}]},{"prim":"storage","args":[{"prim":"key_hash"}]},{"prim":"code","args":[[[[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]],{"prim":"IF_LEFT","args":[[{"prim":"PUSH","args":[{"prim":"mutez"},{"int":"0"}]},{"prim":"AMOUNT"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],[{"prim":"DIP","args":[[{"prim":"DUP"}]]},{"prim":"SWAP"}],{"prim":"IMPLICIT_ACCOUNT"},{"prim":"ADDRESS"},{"prim":"SENDER"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]]}]`
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
