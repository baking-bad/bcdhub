package contractparser

import (
	"regexp"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// FindHardcodedAddresses -
func FindHardcodedAddresses(script gjson.Result) (helpers.Set, error) {
	s := script.String()
	regexString := "(tz|KT)[0-9A-Za-z]{34}"
	re := regexp.MustCompile(regexString)
	res := re.FindAllString(s, -1)
	resp := make(helpers.Set)
	resp.Append(res...)
	return resp, nil
}

// IsLiteral -
func IsLiteral(prim string) bool {
	return helpers.StringInArray(prim, []string{
		consts.CONTRACT, consts.BYTES, consts.ADDRESS, consts.KEYHASH,
		consts.KEY, consts.TIMESTAMP, consts.BOOL, consts.MUTEZ,
		consts.NAT, consts.STRING, consts.INT, consts.SIGNATURE,
	})
}
