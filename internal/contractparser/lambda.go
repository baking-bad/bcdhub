package contractparser

import (
	"encoding/hex"
	"regexp"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/tidwall/gjson"
)

// HasLambda -
func HasLambda(data gjson.Result) bool {
	if data.IsObject() {
		prim := data.Get("prim")
		if prim.Exists() {
			if prim.String() == consts.LAMBDA {
				return true
			}
			args := data.Get("args")
			if args.Exists() {
				if HasLambda(args) {
					return true
				}
			}
		} else {
			bytes := data.Get("bytes")
			if bytes.Exists() {
				return detectLambdaByBytes(bytes.String())
			}
		}
	} else if data.IsArray() {
		for _, item := range data.Array() {
			if HasLambda(item) {
				return true
			}
		}
	}
	return false
}

func detectLambdaByBytes(input string) bool {
	if len(input) < 24 {
		return false
	}
	re := regexp.MustCompile("^0502[0-9a-f]{8}0[3-9]")
	if !re.MatchString(input) {
		return false
	}
	b, err := hex.DecodeString(input[22:24])
	if err != nil {
		logger.Error(err)
		return false
	}
	if len(b) != 1 {
		return false
	}
	if 0x0c > b[0] || 0x75 < b[0] {
		return false
	}
	return 0x58 >= b[0] || 0x6f <= b[0]
}
