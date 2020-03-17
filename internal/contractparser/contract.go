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
